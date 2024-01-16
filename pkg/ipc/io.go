package ipc

import (
	"errors"
	"fmt"
	"path/filepath"
	"tonysoft.com/gothon/internal/code"
	"tonysoft.com/gothon/internal/io"
)

func initIO(gothonDir string, nodeCount int, pkg code.Package) (io.SocketArray, error) {
	socketArray := io.NewSocketArray(filepath.Join(gothonDir, "sock"), getSocketPaths(pkg), nodeCount)
	err := socketArray.Listen()
	if err != nil {
		return nil, err
	}

	return socketArray, nil
}

func getSocketPaths(p code.Package) (paths []string) {
	pathsMap := make(map[string]any)
	invalidErr := errors.New("invalid action for variable type")

	for _, mod := range p {
		for _, stmt := range mod.Statements {
			if stmt.ShouldSkip {
				continue
			}

			if stmt.Actions.Contains(code.VariableDefinition) {
				switch stmt.TargetVariable.Type {
				case code.LockFunc, code.UnlockFunc, code.WaitGroup:
					pathsMap[fmt.Sprintf("%s/%s_%s", mod.Name, stmt.TargetVariable.Name, "in")] = nil
					pathsMap[fmt.Sprintf("%s/%s_%s", mod.Name, stmt.TargetVariable.Name, "out")] = nil
				default:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "set_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "set_out")] = nil
				}
			}

			if stmt.Actions.Contains(code.VariableAssignment) || stmt.Actions.Contains(code.QueuePut) {
				switch stmt.TargetVariable.Type {
				case code.LockFunc, code.UnlockFunc:
					panic(invalidErr)
				default:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "set_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "set_out")] = nil
				}
			}

			if stmt.Actions == code.QueueGet {
				switch stmt.TargetVariable.Type {
				case code.Queue, code.LifoQueue:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "get_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "get_out")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "get_ok")] = nil
				}
			}

			if stmt.Actions.Contains(code.VariableUsage) || stmt.Actions == code.QueueGet {
				for _, v := range stmt.UsedVariables {
					switch v.Type {
					case code.LockFunc, code.UnlockFunc:
						panic(invalidErr)
					case code.Queue, code.LifoQueue:
						pathsMap[filepath.Join(mod.Name, v.Name, "get_in")] = nil
						pathsMap[filepath.Join(mod.Name, v.Name, "get_out")] = nil
						pathsMap[filepath.Join(mod.Name, v.Name, "get_ok")] = nil
					default:
						pathsMap[filepath.Join(mod.Name, v.Name, "get_in")] = nil
						pathsMap[filepath.Join(mod.Name, v.Name, "get_out")] = nil
					}
				}
			}

			if stmt.Actions.Contains(code.VariableAdd) {
				switch stmt.TargetVariable.Type {
				case code.Int, code.Float, code.Str:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "add_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "add_out")] = nil
				default:
					panic(invalidErr)
				}
			}

			if stmt.Actions.Contains(code.VariableSubtract) {
				switch stmt.TargetVariable.Type {
				case code.Int, code.Float, code.Str:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "sub_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "sub_out")] = nil
				default:
					panic(invalidErr)
				}
			}

			if stmt.Actions.Contains(code.VariableMultiply) {
				switch stmt.TargetVariable.Type {
				case code.Int, code.Float:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "mul_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "mul_out")] = nil
				default:
					panic(invalidErr)
				}
			}

			if stmt.Actions.Contains(code.VariableDivide) {
				switch stmt.TargetVariable.Type {
				case code.Int, code.Float:
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "div_in")] = nil
					pathsMap[filepath.Join(mod.Name, stmt.TargetVariable.Name, "div_out")] = nil
				default:
					panic(invalidErr)
				}
			}

			if stmt.Actions.Contains(code.QueueSize) {
				for _, v := range stmt.UsedVariables {
					if v.Type == code.Queue || v.Type == code.LifoQueue {
						pathsMap[filepath.Join(mod.Name, v.Name, "size_in")] = nil
						pathsMap[filepath.Join(mod.Name, v.Name, "size_out")] = nil
					}
				}
			}

			if stmt.Actions.Contains(code.QueueEmpty) {
				for _, v := range stmt.UsedVariables {
					if v.Type == code.Queue || v.Type == code.LifoQueue {
						pathsMap[filepath.Join(mod.Name, v.Name, "empty_in")] = nil
						pathsMap[filepath.Join(mod.Name, v.Name, "empty_out")] = nil
					}
				}
			}

			if stmt.Actions.Contains(code.QueueFull) {
				for _, v := range stmt.UsedVariables {
					if v.Type == code.Queue || v.Type == code.LifoQueue {
						pathsMap[filepath.Join(mod.Name, v.Name, "full_in")] = nil
						pathsMap[filepath.Join(mod.Name, v.Name, "full_out")] = nil
					}
				}
			}
		}
	}

	for k := range pathsMap {
		paths = append(paths, k)
	}

	return paths
}
