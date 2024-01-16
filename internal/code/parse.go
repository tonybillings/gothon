package code

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	variableUsageRequireParensKey = "gothon:var_usage:require_parens"
	variableDefinitionPrefixKey   = "gothon:var_def:prefix"
	variableDefinitionSuffixKey   = "gothon:var_def:suffix"
)

var (
	ignoredLinePrefixes      = []string{"import ", "from ", "class ", "def ", "try:", "except ", "finally:", "else:", "\"\"\"", "#", "@"}
	controlStructurePrefixes = []string{"if ", "if(", "elif ", "elif(", "while ", "while(", "for ", "for("}
	supportedTypes           = []string{"bool", "int", "float", "str", "callable"}
	supportedQueueTypes      = []string{"Queue[bool]", "Queue[int]", "Queue[float]", "Queue[str]"}
	supportedLifoQueueTypes  = []string{"LifoQueue[bool]", "LifoQueue[int]", "LifoQueue[float]", "LifoQueue[str]"}
	supportedTypesCombined   = append(append(supportedTypes, supportedQueueTypes...), supportedLifoQueueTypes...)
)

func Parse(projectDir string) (Package, error) {
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	modules, err := getModules(projectDir)
	if err != nil {
		return nil, err
	}

	err = getVariableDefinitions(modules)
	if err != nil {
		return nil, err
	}

	err = getVariableNumericalOperations(modules)
	if err != nil {
		return nil, err
	}

	err = getQueueOperations(modules)
	if err != nil {
		return nil, err
	}

	err = getVariableAssignments(modules)
	if err != nil {
		return nil, err
	}

	err = getMutexLocksAndUnlocks(modules)
	if err != nil {
		return nil, err
	}

	err = getSyncs(modules)
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func getModules(packageDir string) ([]*Module, error) {
	modules := make([]*Module, 0)
	err := filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(path, filepath.Join(packageDir, ".gothon")) ||
			!strings.HasSuffix(info.Name(), ".py") {
			return nil
		}

		moduleName := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(path, packageDir), ".py"), "/")
		moduleName = strings.ReplaceAll(moduleName, "/", "_")
		relativePath := strings.TrimPrefix(strings.TrimPrefix(path, packageDir), "/")

		module := &Module{
			Name:             moduleName,
			AbsolutePath:     path,
			RelativePath:     relativePath,
			PackageDirectory: packageDir,
			Statements:       make([]*Statement, 0),
		}

		err = processDirectives(module)
		if err != nil {
			return err
		}

		modules = append(modules, module)
		return nil
	})

	return modules, err
}

func processDirectives(module *Module) error {
	moduleFile, err := os.Open(module.AbsolutePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = moduleFile.Close()
	}()

	scanner := bufio.NewScanner(moduleFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, "# gothon:") {
			continue
		}

		if strings.HasPrefix(text, "# "+variableUsageRequireParensKey) {
			kv := strings.Split(text, "=")
			val := strings.TrimSpace(kv[1])
			valBool, e := strconv.ParseBool(val)
			if e != nil {
				return e
			}
			module.RequireParens = valBool
			continue
		}

		if strings.HasPrefix(text, "# "+variableDefinitionPrefixKey) {
			kv := strings.Split(text, "=")
			val := strings.TrimSpace(kv[1])
			module.VariablePrefix = val
			continue
		}

		if strings.HasPrefix(text, "# "+variableDefinitionSuffixKey) {
			kv := strings.Split(text, "=")
			val := strings.TrimSpace(kv[1])
			module.VariableSuffix = val
			continue
		}
	}

	if module.VariablePrefix == "" {
		module.VariablePrefix = "_"
	} else if module.VariablePrefix == "None" {
		module.VariablePrefix = ""
	}

	if module.VariableSuffix == "" {
		module.VariableSuffix = "_"
	} else if module.VariableSuffix == "None" {
		module.VariableSuffix = ""
	}

	return nil
}

func getVariableDefinitions(modules []*Module) error {
	for _, module := range modules {
		f, err := os.Open(module.AbsolutePath)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())

			if shouldSkipLine(text) {
				continue
			}

			textPartsEq := strings.Split(text, "=")
			if len(textPartsEq) < 2 {
				continue
			}

			textPartsCol := strings.Split(textPartsEq[0], ":")
			if len(textPartsCol) != 2 {
				continue
			}

			varType := strings.TrimSpace(textPartsCol[1])
			varTypeSupported := false
			for _, dt := range supportedTypesCombined {
				if varType == dt {
					varTypeSupported = true
					break
				}
			}
			if !varTypeSupported {
				continue
			}

			varname := strings.Replace(strings.TrimSpace(textPartsCol[0]), "self.", "", 1)
			if !strings.HasPrefix(varname, module.VariablePrefix) || !strings.HasSuffix(varname, module.VariableSuffix) {
				continue
			}

			varnameTrimmed := strings.TrimPrefix(varname, module.VariablePrefix)
			varnameTrimmed = strings.TrimSuffix(varnameTrimmed, module.VariableSuffix)
			found, e := regexp.MatchString("[_A-Za-z][_A-Za-z0-9]*", varnameTrimmed)
			if !found || e != nil {
				continue
			}

			shouldSkip := false
			if varname == fmt.Sprintf("%s%s%s", module.VariablePrefix, "node", module.VariableSuffix) {
				shouldSkip = true
			}
			if varname == fmt.Sprintf("%s%s%s", module.VariablePrefix, "node_count", module.VariableSuffix) {
				shouldSkip = true
			}
			if shouldSkip {
				statement := &Statement{
					Line:         line,
					Actions:      VariableDefinition,
					OriginalCode: scanner.Text(),
					ShouldSkip:   true,
				}

				module.Statements = append(module.Statements, statement)
				continue
			}

			if varType == "callable" &&
				(!strings.HasPrefix(varnameTrimmed, "lock_") && !strings.HasPrefix(varnameTrimmed, "unlock_") &&
					!strings.HasPrefix(varnameTrimmed, "sync_")) {
				continue
			}

			tag := ""
			if varType == "callable" {
				t := strings.TrimPrefix(varnameTrimmed, "unlock_")
				t = strings.TrimPrefix(t, "lock_")
				t = fmt.Sprintf("%smutex_%s%s", module.VariablePrefix, t, module.VariableSuffix)
				tag = t
			}

			lValue, rValue := getLRValues(scanner.Text())
			defaultValue, e := getDefaultValue(varType, rValue)
			if e != nil {
				return e
			}

			variable := &Variable{
				ID:           filepath.Join(module.Name, varname),
				Type:         getVariableType(varType, varnameTrimmed),
				SubType:      getVariableSubType(varType),
				Name:         varname,
				Tag:          tag,
				DefaultValue: defaultValue,
			}

			statement := &Statement{
				Line:           line,
				Indentation:    getIndentation(scanner.Text()),
				Actions:        VariableDefinition,
				TargetVariable: variable,
				OriginalCode:   scanner.Text(),
				ModifiedCode:   scanner.Text(),
				OriginalLValue: lValue,
				OriginalRValue: rValue,
				ModifiedRValue: rValue,
			}

			module.Statements = append(module.Statements, statement)
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getVariableAssignments(modules []*Module) error {
	for _, module := range modules {
		f, err := os.Open(module.AbsolutePath)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())

			shouldContinue := false
			for _, prefix := range ignoredLinePrefixes {
				if strings.HasPrefix(text, prefix) {
					shouldContinue = true
					break
				}
			}
			if shouldContinue {
				continue
			}

			lValue, rValue := getLRValues(scanner.Text())

			statement := &Statement{
				Line:           line,
				Indentation:    getIndentation(scanner.Text()),
				UsedVariables:  make([]*Variable, 0),
				OriginalCode:   scanner.Text(),
				ModifiedCode:   scanner.Text(),
				OriginalLValue: lValue,
				OriginalRValue: rValue,
				ModifiedRValue: rValue,
			}

			textParts := strings.Split(text, "=")
			if len(textParts) == 2 {
				varname := getVariableName(textParts[0])

				targetVariable := module.GetVariableByName(varname)
				if targetVariable != nil {
					statement.Actions = VariableAssignment
					statement.TargetVariable = targetVariable

				}

				if module.GetStatement(line) == nil {
					checkForVariableUsage(statement, textParts[1], module.GetVariables(), module.RequireParens)
				}
			} else {
				checkForVariableUsage(statement, text, module.GetVariables(), module.RequireParens)
			}

			if statement.Actions != 0 {
				module.Statements = append(module.Statements, statement)
			}
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getVariableNumericalOperations(modules []*Module) error {
	for _, module := range modules {
		f, err := os.Open(module.AbsolutePath)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())

			shouldContinue := false
			for _, prefix := range ignoredLinePrefixes {
				if strings.HasPrefix(text, prefix) {
					shouldContinue = true
					break
				}
			}
			if shouldContinue {
				continue
			}

			lValue, rValue := getLRValues(scanner.Text())

			statement := &Statement{
				Line:           line,
				Indentation:    getIndentation(scanner.Text()),
				UsedVariables:  make([]*Variable, 0),
				OriginalCode:   scanner.Text(),
				ModifiedCode:   scanner.Text(),
				OriginalLValue: lValue,
				OriginalRValue: rValue,
				ModifiedRValue: rValue,
			}

			textParts := strings.Split(text, "+=")
			if len(textParts) == 2 {
				varname := getVariableName(textParts[0])

				targetVariable := module.GetVariableByName(varname)
				if targetVariable != nil {
					statement.Actions = VariableAdd
					statement.TargetVariable = targetVariable
				}

				checkForVariableUsage(statement, textParts[1], module.GetVariables(), module.RequireParens)
			}

			textParts = strings.Split(text, "-=")
			if len(textParts) == 2 {
				varname := getVariableName(textParts[0])

				targetVariable := module.GetVariableByName(varname)
				if targetVariable != nil {
					statement.Actions = VariableSubtract
					statement.TargetVariable = targetVariable
				}

				checkForVariableUsage(statement, textParts[1], module.GetVariables(), module.RequireParens)
			}

			textParts = strings.Split(text, "*=")
			if len(textParts) == 2 {
				varname := getVariableName(textParts[0])

				targetVariable := module.GetVariableByName(varname)
				if targetVariable != nil {
					statement.Actions = VariableMultiply
					statement.TargetVariable = targetVariable
				}

				checkForVariableUsage(statement, textParts[1], module.GetVariables(), module.RequireParens)
			}

			textParts = strings.Split(text, "/=")
			if len(textParts) == 2 {
				varname := getVariableName(textParts[0])

				targetVariable := module.GetVariableByName(varname)
				if targetVariable != nil {
					statement.Actions = VariableDivide
					statement.TargetVariable = targetVariable
				}

				checkForVariableUsage(statement, textParts[1], module.GetVariables(), module.RequireParens)
			}

			if statement.Actions != 0 {
				module.Statements = append(module.Statements, statement)
			}
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getQueueOperations(modules []*Module) error {
	for _, module := range modules {
		f, err := os.Open(module.AbsolutePath)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())

			shouldContinue := false
			for _, prefix := range ignoredLinePrefixes {
				if strings.HasPrefix(text, prefix) {
					shouldContinue = true
					break
				}
			}
			if shouldContinue {
				continue
			}

			lValue, rValue := getLRValues(scanner.Text())

			statement := &Statement{
				Line:           line,
				Indentation:    getIndentation(scanner.Text()),
				UsedVariables:  make([]*Variable, 0),
				OriginalCode:   scanner.Text(),
				ModifiedCode:   scanner.Text(),
				OriginalLValue: lValue,
				OriginalRValue: rValue,
				ModifiedRValue: rValue,
			}

			for _, v := range module.GetVariables() {
				if v.Type != Queue && v.Type != LifoQueue {
					continue
				}

				appendVar := false

				sizeRegex, e := regexp.Compile("\\(?" + v.Name + "\\)?\\.(q)?size\\(")
				if e != nil {
					return e
				}
				if sizeRegex.MatchString(text) {
					statement.Actions |= QueueSize
					appendVar = true
				}

				emptyRegex, e := regexp.Compile("\\(?" + v.Name + "\\)?\\.empty\\(")
				if e != nil {
					return e
				}
				if emptyRegex.MatchString(text) {
					statement.Actions |= QueueEmpty
					appendVar = true
				}

				fullRegex, e := regexp.Compile("\\(?" + v.Name + "\\)?\\.full\\(")
				if e != nil {
					return e
				}
				if fullRegex.MatchString(text) {
					statement.Actions |= QueueFull
					appendVar = true
				}

				putRegex, e := regexp.Compile("\\(?" + v.Name + "\\)?\\.put\\(")
				if e != nil {
					return e
				}
				if putRegex.MatchString(text) {
					statement.Actions |= QueuePut
					statement.TargetVariable = v
				}

				getRegex, e := regexp.Compile("\\(?" + v.Name + "\\)?\\.get\\(")
				if e != nil {
					return e
				}
				if getRegex.MatchString(text) {
					statement.Actions |= QueueGet
					statement.TargetVariable = v
				}

				if appendVar {
					statement.UsedVariables = append(statement.UsedVariables, v)
				}
			}

			if statement.Actions != 0 {
				module.Statements = append(module.Statements, statement)
			}
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func checkForVariableUsage(statement *Statement, expression string, variables []*Variable, requireParens bool) {
	expression = strings.TrimSpace(expression)
	for _, variable := range variables {
		if variable.Type == LockFunc || variable.Type == UnlockFunc ||
			variable.Type == WaitGroup ||
			variable.Type == Queue || variable.Type == LifoQueue {
			continue
		}

		if requireParens {
			if strings.Contains(expression, fmt.Sprintf("(%s)", variable.Name)) {
				statement.Actions |= VariableUsage
				statement.UsedVariables = append(statement.UsedVariables, variable)
			}
		} else {
			if containsVariable(expression, variable.Name) {
				statement.Actions |= VariableUsage
				statement.UsedVariables = append(statement.UsedVariables, variable)
			}
		}
	}
}

func getMutexLocksAndUnlocks(modules []*Module) error {
	for _, module := range modules {
		f, err := os.Open(module.AbsolutePath)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())

			shouldContinue := false
			for _, prefix := range ignoredLinePrefixes {
				if strings.HasPrefix(text, prefix) {
					shouldContinue = true
					break
				}
			}
			if shouldContinue {
				continue
			}

			lValue, rValue := getLRValues(scanner.Text())

			statement := &Statement{
				Line:           line,
				Indentation:    getIndentation(scanner.Text()),
				Actions:        0,
				UsedVariables:  make([]*Variable, 0),
				OriginalCode:   scanner.Text(),
				ModifiedCode:   scanner.Text(),
				OriginalLValue: lValue,
				OriginalRValue: rValue,
				ModifiedRValue: rValue,
			}

			for _, variable := range module.GetVariables() {
				if containsOnlyVariable(text, fmt.Sprintf("%s()", variable.Name)) {
					if variable.Type == LockFunc {
						statement.Actions = MutexLock
						statement.TargetVariable = variable
						break
					} else if variable.Type == UnlockFunc {
						statement.Actions = MutexUnlock
						statement.TargetVariable = variable
						break
					}
				}
			}

			if statement.Actions != 0 {
				module.Statements = append(module.Statements, statement)
			}
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getSyncs(modules []*Module) error {
	for _, module := range modules {
		f, err := os.Open(module.AbsolutePath)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())

			shouldContinue := false
			for _, prefix := range ignoredLinePrefixes {
				if strings.HasPrefix(text, prefix) {
					shouldContinue = true
					break
				}
			}
			if shouldContinue {
				continue
			}

			lValue, rValue := getLRValues(scanner.Text())

			statement := &Statement{
				Line:           line,
				Indentation:    getIndentation(scanner.Text()),
				Actions:        0,
				UsedVariables:  make([]*Variable, 0),
				OriginalCode:   scanner.Text(),
				ModifiedCode:   scanner.Text(),
				OriginalLValue: lValue,
				OriginalRValue: rValue,
				ModifiedRValue: rValue,
			}

			for _, variable := range module.GetVariables() {
				if startsWithVariable(text, fmt.Sprintf("%s(", variable.Name)) {
					switch variable.Type {
					case WaitGroup:
						statement.Actions = Wait
						statement.TargetVariable = variable
						break
					}
				}
			}

			if statement.Actions != 0 {
				module.Statements = append(module.Statements, statement)
			}
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getVariableType(pythonType string, name string) VariableType {
	switch pythonType {
	case "bool":
		return Bool
	case "int":
		return Int
	case "float":
		return Float
	case "str":
		return Str
	case "callable":
		if strings.HasPrefix(name, "lock_") {
			return LockFunc
		}

		if strings.HasPrefix(name, "unlock_") {
			return UnlockFunc
		}

		if strings.HasPrefix(name, "sync_") {
			return WaitGroup
		}
	}

	if strings.HasPrefix(pythonType, "LifoQueue") {
		return LifoQueue
	}

	if strings.HasPrefix(pythonType, "Queue") {
		return Queue
	}

	return 0
}

func getVariableSubType(pythonType string) VariableType {
	if (strings.HasPrefix(pythonType, "LifoQueue[") || strings.HasPrefix(pythonType, "Queue[")) &&
		strings.HasSuffix(pythonType, "]") {
		varType := pythonType[strings.Index(pythonType, "[")+1 : strings.LastIndex(pythonType, "]")]
		return getVariableType(varType, "")
	}
	return 0
}

func shouldSkipLine(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return true
	}

	for _, prefix := range ignoredLinePrefixes {
		if strings.HasPrefix(code, prefix) {
			return true
		}
	}

	return false
}

func getLRValues(code string) (lValue, rValue string) {
	code = strings.TrimSpace(code)

	if shouldSkipLine(code) {
		return "", ""
	}

	for _, prefix := range controlStructurePrefixes {
		if strings.HasPrefix(code, prefix) {
			return "", code
		}
	}

	isAssignment := true
	for _, c := range code {
		if c == '(' {
			isAssignment = false
			break
		} else if c == '=' {
			break
		}
	}

	if isAssignment {
		codeParts := strings.Split(code, "=")
		if len(codeParts) < 2 {
			return "", code
		}

		return codeParts[0], strings.Join(codeParts[1:], "=")
	}

	return "", code
}

func getDefaultValue(dataType string, rValue string) (any, error) {
	rValue = strings.TrimSpace(rValue)

	switch dataType {
	case "callable":
		valParts := strings.Split(rValue, "=")
		if len(valParts) < 2 {
			return rValue, nil
		}
		valParts = strings.Split(valParts[1], ":")
		val := strings.TrimSpace(valParts[0])
		return val, nil
	case "bool":
		return strconv.ParseBool(rValue)
	case "int":
		return strconv.ParseInt(rValue, 10, 64)
	case "float":
		return strconv.ParseFloat(rValue, 64)
	case "str":
		if strings.HasPrefix(rValue, "'") && strings.HasSuffix(rValue, "'") {
			rValue = strings.Trim(rValue, "'")
		} else if strings.HasPrefix(rValue, "\"") && strings.HasSuffix(rValue, "\"") {
			rValue = strings.Trim(rValue, "\"")
		}
		return rValue, nil
	}

	if strings.HasPrefix(dataType, "LifoQueue") {
		rValue = strings.TrimSpace(rValue)
		rValue = strings.TrimPrefix(rValue, "LifoQueue(")
		rValue = strings.TrimSuffix(rValue, ")")
		rValue = strings.TrimSpace(rValue)
		rValue = strings.ReplaceAll(rValue, "_", "")
		return strconv.ParseInt(rValue, 10, 64)
	}

	if strings.HasPrefix(dataType, "Queue") {
		rValue = strings.TrimSpace(rValue)
		rValue = strings.TrimPrefix(rValue, "Queue(")
		rValue = strings.TrimSuffix(rValue, ")")
		rValue = strings.TrimSpace(rValue)
		rValue = strings.ReplaceAll(rValue, "_", "")
		return strconv.ParseInt(rValue, 10, 64)
	}

	return nil, errors.New("unsupported datatype")
}

func getIndentation(code string) string {
	if strings.TrimSpace(code) == "" {
		return ""
	}

	ind := ""

	if code[0] == ' ' {
		for _, c := range code {
			if c == ' ' {
				ind += " "
			} else {
				break
			}
		}
	} else if code[0] == '\t' {
		for _, c := range code {
			if c == '\t' {
				ind += "\t"
			} else {
				break
			}
		}
	}

	return ind
}

func getVariableName(varname string) string {
	varname = strings.TrimSpace(varname)
	varnameParts := strings.Split(varname, ".")
	return varnameParts[len(varnameParts)-1]
}

func startsWithVariable(code string, variableName string) bool {
	variableName = strings.ReplaceAll(variableName, "(", "\\(")
	variableName = strings.ReplaceAll(variableName, ")", "\\)")
	regex, err := regexp.Compile("^([_A-Za-z][_A-Za-z0-9]*\\.)?(" + variableName + ")")
	if err != nil {
		panic(err)
	}
	return regex.MatchString(code)
}

func containsVariable(code string, variableName string) bool {
	regex, err := regexp.Compile("([_A-Za-z][_A-Za-z0-9]*\\.)?(" + variableName + ")")
	if err != nil {
		panic(err)
	}
	return regex.MatchString(code)
}

func containsOnlyVariable(code string, variableName string) bool {
	regex, err := regexp.Compile("^([_A-Za-z][_A-Za-z0-9]*\\.)?(" + variableName + ")$")
	if err != nil {
		panic(err)
	}
	return regex.MatchString(code)
}
