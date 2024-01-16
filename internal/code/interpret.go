package code

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"tonysoft.com/gothon/internal/memory/config"
)

type SocketModule string

func (s SocketModule) String() string {
	return string(s)
}

func Interpret(pkg Package) (SocketModule, error) {
	err := interpretStatements(pkg)
	if err != nil {
		return "", err
	}

	return getSocketModule(pkg)
}

func interpretStatements(pkg Package) error {
	for _, m := range pkg {
		for _, s := range m.Statements {
			err := interpretStatement(s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func interpretStatement(s *Statement) error {
	if s.Actions.Contains(VariableDefinition) || s.ShouldSkip {
		return nil
	}

	if s.Actions.Contains(VariableAssignment) {
		for _, v := range s.UsedVariables {
			regex, err := regexp.Compile("([_A-Za-z][_A-Za-z0-9]*\\.)?(" + v.Name + ")")
			if err != nil {
				panic(err)
			}

			get := getFuncCall(v.ID, "get")
			s.ModifiedRValue = strings.TrimSpace(regex.ReplaceAllString(s.OriginalRValue, get))
		}

		s.ModifiedRValue = getFuncCall(s.TargetVariable.ID, "set", s.ModifiedRValue)
		s.ModifiedCode = fmt.Sprintf("%s%s= %s", s.Indentation, s.OriginalLValue, s.ModifiedRValue)
		return nil
	}

	if s.Actions.Contains(MutexLock) || s.Actions.Contains(MutexUnlock) {
		regex, err := regexp.Compile("([_A-Za-z][_A-Za-z0-9]*\\.)?(" + s.TargetVariable.Name + ")")
		if err != nil {
			panic(err)
		}

		s.ModifiedRValue = regex.ReplaceAllString(s.OriginalRValue, getFuncCall(s.TargetVariable.ID, "mutex", s.OriginalRValue))
		s.ModifiedCode = s.Indentation + s.ModifiedRValue
		return nil
	}

	if s.Actions.Contains(Wait) {
		regex, err := regexp.Compile("([_A-Za-z][_A-Za-z0-9]*\\.)?(" + s.TargetVariable.Name + ")\\([0-9]*\\)")
		if err != nil {
			panic(err)
		}

		syncCount := s.OriginalRValue[strings.Index(s.OriginalRValue, "(")+1 : strings.LastIndex(s.OriginalRValue, ")")]
		s.ModifiedRValue = regex.ReplaceAllString(s.OriginalRValue, getFuncCall(s.TargetVariable.ID, "sync", syncCount))
		s.ModifiedCode = s.Indentation + s.ModifiedRValue
		return nil
	}

	if s.Actions.Contains(VariableUsage) {
		for _, v := range s.UsedVariables {
			regex, err := regexp.Compile("([_A-Za-z][_A-Za-z0-9]*\\.)?(" + v.Name + ")")
			if err != nil {
				panic(err)
			}

			get := getFuncCall(v.ID, "get")
			s.ModifiedRValue = strings.TrimSpace(regex.ReplaceAllString(s.ModifiedRValue, get))
		}

		if s.Actions == VariableUsage && s.OriginalLValue == "" {
			s.ModifiedCode = s.Indentation + s.ModifiedRValue
			return nil
		}
	}

	if s.Actions.Contains(VariableAdd) {
		s.ModifiedRValue = getFuncCall(s.TargetVariable.ID, "add", s.ModifiedRValue)
	}

	if s.Actions.Contains(VariableSubtract) {
		s.ModifiedRValue = getFuncCall(s.TargetVariable.ID, "sub", s.ModifiedRValue)

		if s.TargetVariable.Type == Str {
			s.ModifiedCode = fmt.Sprintf("%s%s= %s", s.Indentation, s.OriginalLValue, strings.TrimSpace(s.ModifiedRValue))
			s.ModifiedCode = strings.Replace(s.ModifiedCode, "-=", "=", 1)
			return nil
		}
	}

	if s.Actions.Contains(VariableMultiply) {
		s.ModifiedRValue = getFuncCall(s.TargetVariable.ID, "mul", s.ModifiedRValue)
	}

	if s.Actions.Contains(VariableDivide) {
		s.ModifiedRValue = getFuncCall(s.TargetVariable.ID, "div", s.ModifiedRValue)
	}

	subType := ""
	if s.TargetVariable != nil {
		subType = fmt.Sprintf("%s_", s.TargetVariable.SubType)
	}

	if s.Actions.Contains(QueueSize) {
		for _, v := range s.UsedVariables {
			s.ModifiedRValue = getQueueFuncCall(s.OriginalRValue, v.Name, v.ID, "size")
		}

		if s.Actions == QueueSize {
			s.ModifiedCode = s.Indentation + s.ModifiedRValue
			return nil
		}
	}

	if s.Actions.Contains(QueueEmpty) {
		for _, v := range s.UsedVariables {
			s.ModifiedRValue = getQueueFuncCall(s.OriginalRValue, v.Name, v.ID, "empty")
		}

		if s.Actions == QueueEmpty {
			s.ModifiedCode = s.Indentation + s.ModifiedRValue
			return nil
		}
	}

	if s.Actions.Contains(QueueFull) {
		for _, v := range s.UsedVariables {
			s.ModifiedRValue = getQueueFuncCall(s.OriginalRValue, v.Name, v.ID, "full")
		}

		if s.Actions == QueueFull {
			s.ModifiedCode = s.Indentation + s.ModifiedRValue
			return nil
		}
	}

	if s.Actions.Contains(QueuePut) {
		arg := strings.TrimSpace(s.ModifiedRValue)
		arg = arg[strings.Index(arg, "(")+1 : strings.LastIndex(arg, ")")]
		s.ModifiedRValue = getQueueFuncCall(s.OriginalRValue, s.TargetVariable.Name, s.TargetVariable.ID, subType+"put", arg)
	}

	if s.Actions == QueueGet {
		s.ModifiedRValue = getQueueFuncCall(s.OriginalRValue, s.TargetVariable.Name, s.TargetVariable.ID, subType+"get")
	}

	if s.ModifiedRValue == s.OriginalRValue {
		return errors.New("interpretation error: ModifiedRValue == OriginalRValue")
	}

	s.ModifiedCode = fmt.Sprintf("%s%s= %s", s.Indentation, s.OriginalLValue, strings.TrimSpace(s.ModifiedRValue))
	return nil
}

func getFuncDefinition(variableID string, varType VariableType, action string) (name, def string) {
	switch {
	case action == "mutex", action == "sync":
		name = fmt.Sprintf("gothon_%s", translateID(variableID))
		def = fillTemplate(templates[action], translateID(variableID), action)
		return name, def
	case strings.HasPrefix(action, "queue_"):
		queueAction := strings.TrimPrefix(action, "queue_")
		name = fmt.Sprintf("gothon_%s_%s", translateID(variableID), queueAction)
		def = fillTemplate(templates[action], translateID(variableID), queueAction)
		return name, def
	case strings.Contains(action, "_queue_"):
		actionParts := strings.Split(action, "_")
		queueAction := actionParts[2]
		name = fmt.Sprintf("gothon_%s_%s", translateID(variableID), queueAction)
		def = fillTemplate(templates[action], translateID(variableID), queueAction)
		return name, def
	default:
		name = fmt.Sprintf("gothon_%s_%s", translateID(variableID), action)
		def = fillTemplate(templates[varType.String()+"_"+action], translateID(variableID), action)
		return name, def
	}
}

func getFuncCall(variableID string, action string, arg ...string) string {
	variableID = translateID(variableID)
	switch action {
	case "mutex":
		return fmt.Sprintf("gothon_%s()", variableID)
	case "sync":
		return fmt.Sprintf("gothon_%s(%s)", variableID, arg[0])
	case "get", "size", "empty", "full":
		return fmt.Sprintf("gothon_%s_%s()", variableID, action)
	default:
		actionParts := strings.Split(action, "_")
		if len(actionParts) > 1 {
			switch actionParts[1] {
			case "put":
				return fmt.Sprintf("gothon_%s_set(", variableID)
			case "get":
				return fmt.Sprintf("gothon_%s_get()", variableID)
			}
		} else {
			return fmt.Sprintf("gothon_%s_%s(%s)", variableID, action, strings.TrimSpace(arg[0]))
		}
	}

	return ""
}

func getQueueFuncCall(rValue string, variableName string, variableID string, action string, arg ...string) string {
	sizeRegex, e := regexp.Compile("\\(?" + variableName + "\\)?\\.(q)?size\\(")
	if e != nil {
		panic(e)
	}
	rValue = sizeRegex.ReplaceAllString(rValue, strings.TrimSuffix(getFuncCall(variableID, action, arg...), ")"))

	emptyRegex, e := regexp.Compile("\\(?" + variableName + "\\)?\\.empty\\(")
	if e != nil {
		panic(e)
	}
	rValue = emptyRegex.ReplaceAllString(rValue, strings.TrimSuffix(getFuncCall(variableID, action, arg...), ")"))

	fullRegex, e := regexp.Compile("\\(?" + variableName + "\\)?\\.full\\(")
	if e != nil {
		panic(e)
	}
	rValue = fullRegex.ReplaceAllString(rValue, strings.TrimSuffix(getFuncCall(variableID, action, arg...), ")"))

	putRegex, e := regexp.Compile("\\(?" + variableName + "\\)?\\.put\\(")
	if e != nil {
		panic(e)
	}
	rValue = putRegex.ReplaceAllString(rValue, strings.TrimSuffix(getFuncCall(variableID, action, arg...), ")"))

	getRegex, e := regexp.Compile("\\(?" + variableName + "\\)?\\.get\\(")
	if e != nil {
		panic(e)
	}
	rValue = getRegex.ReplaceAllString(rValue, strings.TrimSuffix(getFuncCall(variableID, action, arg...), ")"))

	return rValue
}

func getSocketInit(variableID string, action string) (name, code string) {
	switch action {
	case "mutex":
		name = fmt.Sprintf("%s", translateID(variableID))
		code = fillTemplate(socketInitTemplateForMutex, translateID(variableID), "")
		return name, code
	case "sync":
		name = fmt.Sprintf("%s", translateID(variableID))
		code = fillTemplate(socketInitTemplateForSync, translateID(variableID), "")
		return name, code
	case "queue_get":
		name = fmt.Sprintf("%s", translateID(variableID))
		code = fillTemplate(socketInitTemplateForQueueGet, translateID(variableID), "")
		return name, code
	default:
		name = fmt.Sprintf("%s_%s", translateID(variableID), action)
		code = fillTemplate(socketInitTemplate, translateID(variableID), action)
		return name, code
	}
}

func fillTemplate(template string, variableID string, action string) string {
	result := strings.ReplaceAll(template, "{{var_id}}", variableID)
	result = strings.ReplaceAll(result, "{{action}}", action)
	result = strings.ReplaceAll(result, "{{str_max_size}}", fmt.Sprintf("%d", config.GetStringRegisterBufferSize()))
	return result
}

func translateID(variableID string) string {
	variableID = strings.ReplaceAll(variableID, "/", "_")
	variableID = strings.ReplaceAll(variableID, ".", "_")
	return variableID
}

func getSocketModule(pkg Package) (SocketModule, error) {
	sb := strings.Builder{}

	sb.WriteString("import struct\n")
	sb.WriteString("import sys\n")
	sb.WriteString("import socket\n\n")

	socks, addrs, funcs, init, err := getModuleParts(pkg)
	if err != nil {
		return "", err
	}

	writeSocketDefinitions(socks, &sb)
	writeAddressDefinitions(addrs, &sb)
	writeFunctionDefinitions(funcs, &sb)
	writeSocketInit(init, &sb)

	return SocketModule(sb.String()), nil
}

func getModuleParts(pkg Package) (socks, addrs, funcs, init map[string]string, err error) {
	socks = make(map[string]string)
	addrs = make(map[string]string)
	funcs = make(map[string]string)
	init = make(map[string]string)

	for _, m := range pkg {
		for _, s := range m.Statements {
			setSocketDefinitions(socks, s)
			setAddressDefinitions(addrs, s)
			setFunctionDefinitions(funcs, s)
			setSocketInit(init, s)
		}
	}

	if len(socks) != len(addrs) {
		return nil, nil, nil, nil, errors.New("interpretation error: len(socks) != len(addrs)")
	}

	return socks, addrs, funcs, init, nil
}

func setSocketDefinitions(defs map[string]string, s *Statement) {
	getDef := func(name string) string {
		return fmt.Sprintf("%s = socket.socket(socket.AF_UNIX, socket.SOCK_DGRAM)", name)
	}

	if s.ShouldSkip {
		return
	}

	if s.Actions.Contains(VariableDefinition) || s.Actions.Contains(VariableAssignment) {
		if s.TargetVariable.Type == LockFunc || s.TargetVariable.Type == UnlockFunc || s.TargetVariable.Type == WaitGroup {
			name := fmt.Sprintf("_sock_%s_in", translateID(s.TargetVariable.ID))
			defs[name] = getDef(name)

			name = fmt.Sprintf("_sock_%s_out", translateID(s.TargetVariable.ID))
			defs[name] = getDef(name)
		} else {
			name := fmt.Sprintf("_sock_%s_set_in", translateID(s.TargetVariable.ID))
			defs[name] = getDef(name)

			name = fmt.Sprintf("_sock_%s_set_out", translateID(s.TargetVariable.ID))
			defs[name] = getDef(name)
		}
	}

	if s.Actions.Contains(VariableUsage) {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_sock_%s_get_in", translateID(v.ID))
			defs[name] = getDef(name)

			name = fmt.Sprintf("_sock_%s_get_out", translateID(v.ID))
			defs[name] = getDef(name)
		}
	}

	if s.Actions.Contains(VariableAdd) {
		name := fmt.Sprintf("_sock_%s_add_in", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_add_out", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)
	}

	if s.Actions.Contains(VariableSubtract) {
		name := fmt.Sprintf("_sock_%s_sub_in", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_sub_out", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)
	}

	if s.Actions.Contains(VariableMultiply) {
		name := fmt.Sprintf("_sock_%s_mul_in", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_mul_out", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)
	}

	if s.Actions.Contains(VariableDivide) {
		name := fmt.Sprintf("_sock_%s_div_in", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_div_out", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)
	}

	if s.Actions == QueueSize {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_sock_%s_size_in", translateID(v.ID))
			defs[name] = getDef(name)

			name = fmt.Sprintf("_sock_%s_size_out", translateID(v.ID))
			defs[name] = getDef(name)
		}
	}

	if s.Actions == QueueEmpty {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_sock_%s_empty_in", translateID(v.ID))
			defs[name] = getDef(name)

			name = fmt.Sprintf("_sock_%s_empty_out", translateID(v.ID))
			defs[name] = getDef(name)
		}
	}

	if s.Actions == QueueFull {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_sock_%s_full_in", translateID(v.ID))
			defs[name] = getDef(name)

			name = fmt.Sprintf("_sock_%s_full_out", translateID(v.ID))
			defs[name] = getDef(name)
		}
	}

	if s.Actions.Contains(QueuePut) {
		name := fmt.Sprintf("_sock_%s_set_in", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_set_out", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)
	}

	if s.Actions == QueueGet {
		name := fmt.Sprintf("_sock_%s_get_in", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_get_out", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)

		name = fmt.Sprintf("_sock_%s_get_ok", translateID(s.TargetVariable.ID))
		defs[name] = getDef(name)
	}
}

func setAddressDefinitions(addrs map[string]string, s *Statement) {
	getDef := func(name, id, action string) string {
		return fmt.Sprintf("%s = '{{gothon_dir}}/sock/{{node_id}}/%s/%s'", name, id, action)
	}

	if s.ShouldSkip {
		return
	}

	if s.Actions.Contains(VariableDefinition) || s.Actions.Contains(VariableAssignment) || s.Actions.Contains(QueuePut) {
		if s.TargetVariable.Type == LockFunc || s.TargetVariable.Type == UnlockFunc || s.TargetVariable.Type == WaitGroup {
			name := fmt.Sprintf("_addr_%s_in", translateID(s.TargetVariable.ID))
			addrs[name] = fmt.Sprintf("%s = '{{gothon_dir}}/sock/{{node_id}}/%s_in'", name, s.TargetVariable.ID)

			name = fmt.Sprintf("_addr_%s_out", translateID(s.TargetVariable.ID))
			addrs[name] = fmt.Sprintf("%s = '{{gothon_dir}}/sock/{{node_id}}/%s_out'", name, s.TargetVariable.ID)
		} else {
			name := fmt.Sprintf("_addr_%s_set_in", translateID(s.TargetVariable.ID))
			addrs[name] = getDef(name, s.TargetVariable.ID, "set_in")

			name = fmt.Sprintf("_addr_%s_set_out", translateID(s.TargetVariable.ID))
			addrs[name] = getDef(name, s.TargetVariable.ID, "set_out")
		}
	}

	if s.Actions.Contains(VariableUsage) {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_addr_%s_get_in", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "get_in")

			name = fmt.Sprintf("_addr_%s_get_out", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "get_out")
		}
	}

	if s.Actions.Contains(VariableAdd) {
		name := fmt.Sprintf("_addr_%s_add_in", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "add_in")

		name = fmt.Sprintf("_addr_%s_add_out", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "add_out")
	}

	if s.Actions.Contains(VariableSubtract) {
		name := fmt.Sprintf("_addr_%s_sub_in", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "sub_in")

		name = fmt.Sprintf("_addr_%s_sub_out", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "sub_out")
	}

	if s.Actions.Contains(VariableMultiply) {
		name := fmt.Sprintf("_addr_%s_mul_in", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "mul_in")

		name = fmt.Sprintf("_addr_%s_mul_out", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "mul_out")
	}

	if s.Actions.Contains(VariableDivide) {
		name := fmt.Sprintf("_addr_%s_div_in", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "div_in")

		name = fmt.Sprintf("_addr_%s_div_out", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "div_out")
	}

	if s.Actions.Contains(QueueSize) {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_addr_%s_size_in", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "size_in")

			name = fmt.Sprintf("_addr_%s_size_out", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "size_out")
		}
	}

	if s.Actions.Contains(QueueEmpty) {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_addr_%s_empty_in", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "empty_in")

			name = fmt.Sprintf("_addr_%s_empty_out", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "empty_out")
		}
	}

	if s.Actions.Contains(QueueFull) {
		for _, v := range s.UsedVariables {
			name := fmt.Sprintf("_addr_%s_full_in", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "full_in")

			name = fmt.Sprintf("_addr_%s_full_out", translateID(v.ID))
			addrs[name] = getDef(name, v.ID, "full_out")
		}
	}

	if s.Actions == QueueGet {
		name := fmt.Sprintf("_addr_%s_get_in", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "get_in")

		name = fmt.Sprintf("_addr_%s_get_out", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "get_out")

		name = fmt.Sprintf("_addr_%s_get_ok", translateID(s.TargetVariable.ID))
		addrs[name] = getDef(name, s.TargetVariable.ID, "get_ok")
	}
}

func setFunctionDefinitions(funcs map[string]string, s *Statement) {
	var name, def string

	if s.ShouldSkip {
		return
	}

	if s.Actions.Contains(VariableDefinition) || s.Actions.Contains(VariableAssignment) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "set")
	}

	if s.Actions.Contains(QueuePut) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.SubType, fmt.Sprintf("%s_queue_set", s.TargetVariable.SubType))
	}

	if s.Actions.Contains(MutexLock) || s.Actions.Contains(MutexUnlock) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "mutex")
	}

	if s.Actions.Contains(Wait) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "sync")
	}

	if s.Actions.Contains(VariableUsage) {
		for _, v := range s.UsedVariables {
			name, def = getFuncDefinition(v.ID, v.Type, "get")
			funcs[name] = def
		}
	}

	if s.Actions.Contains(VariableAdd) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "add")
	}

	if s.Actions.Contains(VariableSubtract) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "sub")
	}

	if s.Actions.Contains(VariableMultiply) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "mul")
	}

	if s.Actions.Contains(VariableDivide) {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.Type, "div")
	}

	if s.Actions.Contains(QueueSize) {
		for _, v := range s.UsedVariables {
			if v.Type == Queue || v.Type == LifoQueue {
				name, def = getFuncDefinition(v.ID, v.SubType, "queue_size")
				funcs[name] = def
			}
		}
	}

	if s.Actions.Contains(QueueEmpty) {
		for _, v := range s.UsedVariables {
			if v.Type == Queue || v.Type == LifoQueue {
				name, def = getFuncDefinition(v.ID, v.SubType, "queue_empty")
				funcs[name] = def
			}
		}
	}

	if s.Actions.Contains(QueueFull) {
		for _, v := range s.UsedVariables {
			if v.Type == Queue || v.Type == LifoQueue {
				name, def = getFuncDefinition(v.ID, v.SubType, "queue_full")
				funcs[name] = def
			}
		}
	}

	if s.Actions == QueueGet {
		name, def = getFuncDefinition(s.TargetVariable.ID, s.TargetVariable.SubType, fmt.Sprintf("%s_queue_get", s.TargetVariable.SubType))
	}

	funcs[name] = def
}

func setSocketInit(init map[string]string, s *Statement) {
	var name, code string

	if s.ShouldSkip {
		return
	}

	if s.Actions.Contains(VariableDefinition) || s.Actions.Contains(VariableAssignment) || s.Actions.Contains(QueuePut) {
		if s.TargetVariable.Type == LockFunc || s.TargetVariable.Type == UnlockFunc {
			name, code = getSocketInit(s.TargetVariable.ID, "mutex")
		} else if s.TargetVariable.Type == WaitGroup {
			name, code = getSocketInit(s.TargetVariable.ID, "sync")
		} else {
			name, code = getSocketInit(s.TargetVariable.ID, "set")
		}
	}

	if s.Actions.Contains(VariableUsage) {
		for _, v := range s.UsedVariables {
			name, code = getSocketInit(v.ID, "get")
			init[name] = code
		}
	}

	if s.Actions == QueueGet {
		name, code = getSocketInit(s.TargetVariable.ID, "get")
		init[name] = code

		name, code = getSocketInit(s.TargetVariable.ID, "queue_get")
		init[name] = code
	}

	if s.Actions.Contains(VariableAdd) {
		name, code = getSocketInit(s.TargetVariable.ID, "add")
	}

	if s.Actions.Contains(VariableSubtract) {
		name, code = getSocketInit(s.TargetVariable.ID, "sub")
	}

	if s.Actions.Contains(VariableMultiply) {
		name, code = getSocketInit(s.TargetVariable.ID, "mul")
	}

	if s.Actions.Contains(VariableDivide) {
		name, code = getSocketInit(s.TargetVariable.ID, "div")
	}

	if s.Actions.Contains(QueueSize) {
		for _, v := range s.UsedVariables {
			if v.Type == Queue || v.Type == LifoQueue {
				name, code = getSocketInit(v.ID, "size")
				if name != "" && code != "" {
					init[name] = code
				}
			}
		}
	}

	if s.Actions.Contains(QueueEmpty) {
		for _, v := range s.UsedVariables {
			if v.Type == Queue || v.Type == LifoQueue {
				name, code = getSocketInit(v.ID, "empty")
				if name != "" && code != "" {
					init[name] = code
				}
			}
		}
	}

	if s.Actions.Contains(QueueFull) {
		for _, v := range s.UsedVariables {
			if v.Type == Queue || v.Type == LifoQueue {
				name, code = getSocketInit(v.ID, "full")
				if name != "" && code != "" {
					init[name] = code
				}
			}
		}
	}

	if name != "" && code != "" {
		init[name] = code
	}
}

func writeSocketDefinitions(socks map[string]string, sb *strings.Builder) {
	for _, s := range socks {
		sb.WriteString(s + "\n")
	}
	sb.WriteString("\n")
}

func writeAddressDefinitions(addrs map[string]string, sb *strings.Builder) {
	for _, a := range addrs {
		sb.WriteString(a + "\n")
	}
	sb.WriteString("\n")
}

func writeFunctionDefinitions(funcs map[string]string, sb *strings.Builder) {
	for _, f := range funcs {
		sb.WriteString(f + "\n\n")
	}
	sb.WriteString("\n")
}

func writeSocketInit(init map[string]string, sb *strings.Builder) {
	if len(init) == 0 {
		return
	}

	sb.WriteString("try:")

	for _, i := range init {
		sb.WriteString(i)
	}

	sb.WriteString("\nexcept socket.error as msg:\n    print(msg, file=sys.stderr)\n    sys.exit(1)\n")
}
