package ipc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"tonysoft.com/gothon/internal/code"
	"tonysoft.com/gothon/internal/io"
	"tonysoft.com/gothon/internal/memory"
	"tonysoft.com/gothon/internal/memory/config"
	"tonysoft.com/gothon/internal/queue"
)

func initMemory(socketArray io.SocketArray, pkg code.Package, nodeCount int) error {
	err := configureGlobalOptions()
	if err != nil {
		return err
	}

	registry := memory.NewRegistry(getRegisters(pkg, nodeCount))
	configureRegistry(registry, socketArray)
	registry.Init()
	return nil
}

func configureGlobalOptions() error {
	strMaxSize := os.Getenv("GOTHON_STRING_MAX_SIZE")
	if strMaxSize != "" {
		maxSize, err := strconv.Atoi(strMaxSize)
		if err != nil {
			return err
		}
		config.SetStringRegisterBufferSize(uint32(maxSize))
	}
	return nil
}

func getRegisters(pkg code.Package, nodeCount int) []memory.Register {
	regMap := make(map[string]memory.Register)

	for _, mod := range pkg {
		for _, stmt := range mod.Statements {
			if stmt.Actions.Contains(code.VariableDefinition) {
				if stmt.ShouldSkip {
					continue
				}

				switch stmt.TargetVariable.Type {
				case code.Bool:
					regMap[stmt.TargetVariable.ID] = memory.NewRegister[bool](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
				case code.Int:
					regMap[stmt.TargetVariable.ID] = memory.NewRegister[int64](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
				case code.Float:
					regMap[stmt.TargetVariable.ID] = memory.NewRegister[float64](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
				case code.Str:
					regMap[stmt.TargetVariable.ID] = memory.NewRegister[string](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
				case code.LockFunc, code.UnlockFunc:
					pathParts := strings.Split(stmt.TargetVariable.ID, "/")
					action := pathParts[len(pathParts)-1]
					id := strings.TrimSuffix(stmt.TargetVariable.ID, action)
					id += stmt.TargetVariable.Tag
					regMap[id] = memory.NewRegister[sync.Mutex](id, sync.Mutex{})
				case code.WaitGroup:
					defaultVal := 0
					if stmt.TargetVariable.DefaultValue == fmt.Sprintf("%snode_count%s", mod.VariablePrefix, mod.VariableSuffix) {
						defaultVal = nodeCount
					} else {
						v, err := strconv.Atoi(stmt.TargetVariable.DefaultValue.(string))
						if err != nil {
							panic(err)
						}
						defaultVal = v
					}

					var wg sync.WaitGroup
					wg.Add(defaultVal)
					regMap[stmt.TargetVariable.ID] = memory.NewRegister[*sync.WaitGroup](stmt.TargetVariable.ID, &wg)
				case code.Queue:
					switch stmt.TargetVariable.SubType {
					case code.Bool:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Fifo[bool]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					case code.Int:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Fifo[int64]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					case code.Float:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Fifo[float64]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					case code.Str:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Fifo[string]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					}
				case code.LifoQueue:
					switch stmt.TargetVariable.SubType {
					case code.Bool:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Lifo[bool]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					case code.Int:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Lifo[int64]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					case code.Float:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Lifo[float64]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					case code.Str:
						regMap[stmt.TargetVariable.ID] = memory.NewRegister[queue.Lifo[string]](stmt.TargetVariable.ID, stmt.TargetVariable.DefaultValue)
					}
				}

			}
		}
	}

	regs := make([]memory.Register, 0)
	for _, reg := range regMap {
		regs = append(regs, reg)
	}
	return regs
}

func configureRegistry(registry memory.Registry, socketArray io.SocketArray) {
	for _, socket := range socketArray {
		pathParts := strings.Split(socket.Path(), "/")
		action := pathParts[len(pathParts)-1]
		varId := strings.TrimSuffix(socket.Tag, "/"+action)

		switch action {
		case "set_in":
			registry[varId].AddSetterIn(socket)
		case "set_out":
			registry[varId].AddSetterOut(socket)
		case "get_in":
			registry[varId].AddGetterIn(socket)
		case "get_out":
			registry[varId].AddGetterOut(socket)
		case "get_ok":
			registry[varId].AddGetterOk(socket)
		case "add_in":
			registry[varId].AddAdderIn(socket)
		case "add_out":
			registry[varId].AddAdderOut(socket)
		case "sub_in":
			registry[varId].AddSubtractorIn(socket)
		case "sub_out":
			registry[varId].AddSubtractorOut(socket)
		case "mul_in":
			registry[varId].AddMultiplierIn(socket)
		case "mul_out":
			registry[varId].AddMultiplierOut(socket)
		case "div_in":
			registry[varId].AddDividerIn(socket)
		case "div_out":
			registry[varId].AddDividerOut(socket)
		case "size_in":
			registry[varId].AddSizeCallerIn(socket)
		case "size_out":
			registry[varId].AddSizeCallerOut(socket)
		case "empty_in":
			registry[varId].AddEmptyCallerIn(socket)
		case "empty_out":
			registry[varId].AddEmptyCallerOut(socket)
		case "full_in":
			registry[varId].AddFullCallerIn(socket)
		case "full_out":
			registry[varId].AddFullCallerOut(socket)
		default:
			if strings.Contains(socket.Tag, "sync_") {
				varId = socket.Tag
				varId = strings.TrimSuffix(varId, "_in")
				varId = strings.TrimSuffix(varId, "_out")
				if strings.HasSuffix(socket.Tag, "_in") {
					registry[varId].AddSetterIn(socket)
				} else if strings.HasSuffix(socket.Tag, "_out") {
					registry[varId].AddSetterOut(socket)
				}
			} else {
				varId = socket.Tag
				varId = strings.Replace(varId, "unlock_", "mutex_", 1)
				varId = strings.Replace(varId, "unlock_", "mutex_", 1)
				varId = strings.Replace(varId, "lock_", "mutex_", 1)
				varId = strings.Replace(varId, "lock_", "mutex_", 1)
				varId = strings.TrimSuffix(varId, "_in")
				varId = strings.TrimSuffix(varId, "_out")

				if strings.Contains(socket.Tag, "unlock_") && strings.HasSuffix(socket.Tag, "_in") {
					registry[varId].AddUnlockerIn(socket)
				} else if strings.Contains(socket.Tag, "unlock_") && strings.HasSuffix(socket.Tag, "_out") {
					registry[varId].AddUnlockerOut(socket)
				} else if strings.Contains(socket.Tag, "lock_") && strings.HasSuffix(socket.Tag, "_in") {
					registry[varId].AddLockerIn(socket)
				} else if strings.Contains(socket.Tag, "lock_") && strings.HasSuffix(socket.Tag, "_out") {
					registry[varId].AddLockerOut(socket)
				}
			}
		}
	}
}
