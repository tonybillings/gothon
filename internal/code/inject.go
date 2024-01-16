package code

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Inject(pkg Package, socketModule SocketModule, nodeCount int) error {
	err := injectSocketModule(pkg, socketModule, nodeCount)
	if err != nil {
		return err
	}

	return injectModifiedCode(pkg, nodeCount)
}

func injectSocketModule(pkg Package, socketModule SocketModule, nodeCount int) error {
	gothonDir := filepath.Join(pkg.Directory(), ".gothon")

	for i := 0; i < nodeCount; i++ {
		code := strings.ReplaceAll(socketModule.String(), "{{gothon_dir}}", gothonDir)
		code = strings.ReplaceAll(code, "{{node_id}}", strconv.Itoa(i))
		srcDir := filepath.Join(gothonDir, "src", strconv.Itoa(i))
		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				e := os.WriteFile(filepath.Join(path, "_gothon_.py"), []byte(code), 0775)
				if e != nil {
					return e
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func injectModifiedCode(pkg Package, nodeCount int) error {
	gothonDir := filepath.Join(pkg.Directory(), ".gothon")
	srcRootDir := filepath.Join(gothonDir, "src")

	for i := 0; i < nodeCount; i++ {
		srcDir := filepath.Join(srcRootDir, strconv.Itoa(i))

		for _, m := range pkg {
			modulePath := filepath.Join(srcDir, m.RelativePath)
			modifiedCode := strings.Builder{}

			modifiedCode.WriteString("from _gothon_ import *\n\n\n")

			modifiedCode.WriteString(fmt.Sprintf("%snode_count%s: int = %d\n", m.VariablePrefix, m.VariableSuffix, nodeCount))
			modifiedCode.WriteString(fmt.Sprintf("%snode%s: int = %d\n\n\n", m.VariablePrefix, m.VariableSuffix, i))

			moduleFile, err := os.Open(modulePath)
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(moduleFile)
			scanner.Split(bufio.ScanLines)
			line := 0
			for scanner.Scan() {
				line++
				text := scanner.Text()
				stmt := m.GetStatement(line)

				if stmt == nil {
					modifiedCode.WriteString(fmt.Sprintf("%s\n", text))
				} else if !stmt.ShouldSkip {
					modifiedCode.WriteString(fmt.Sprintf("%s\n", stmt.ModifiedCode))
				}
			}

			err = moduleFile.Close()
			if err != nil {
				return err
			}

			err = os.WriteFile(modulePath, []byte(modifiedCode.String()), 0775)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
