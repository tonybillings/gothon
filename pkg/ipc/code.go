package ipc

import (
	"tonysoft.com/gothon/internal/code"
)

func initCode(projectDir string, nodeCount int) (code.Package, error) {
	pkg, err := code.Parse(projectDir)
	if err != nil {
		return nil, err
	}

	socketModule, err := code.Interpret(pkg)
	if err != nil {
		return nil, err
	}

	err = code.Inject(pkg, socketModule, nodeCount)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}
