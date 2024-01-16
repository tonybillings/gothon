package ipc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func Start(ctx context.Context, cancel context.CancelFunc, projectDir string, nodeCount int, nodeArgs string) error {
	gothonDir, err := initSession(projectDir, nodeCount)
	if err != nil {
		return err
	}

	pkg, err := initCode(projectDir, nodeCount)
	if err != nil {
		return err
	}

	socketArray, err := initIO(gothonDir, nodeCount, pkg)
	if err != nil {
		return err
	}

	err = initMemory(socketArray, pkg, nodeCount)
	if err != nil {
		return err
	}

	processGroup := initProcessGroup(gothonDir, nodeCount, nodeArgs, cancel)

	go func() {
		<-ctx.Done()
		processGroup.Stop()
		socketArray.Close()
		closeSession(gothonDir)
	}()

	return nil
}

func initSession(projectDir string, nodeCount int) (gothonDir string, err error) {
	projectDir, err = filepath.Abs(projectDir)
	if err != nil {
		return "", err
	}
	gothonDir = filepath.Join(projectDir, ".gothon")

	_ = os.RemoveAll(gothonDir)
	err = os.MkdirAll(gothonDir, 0775)
	if err != nil {
		return "", err
	}

	sockRootDir := filepath.Join(gothonDir, "sock")
	srcRootDir := filepath.Join(gothonDir, "src")

	for i := 0; i < nodeCount; i++ {
		sockDir := filepath.Join(sockRootDir, strconv.Itoa(i))
		srcDir := filepath.Join(srcRootDir, strconv.Itoa(i))

		err = os.MkdirAll(sockDir, 0775)
		if err != nil {
			return "", err
		}

		err = os.MkdirAll(srcDir, 0775)
		if err != nil {
			return "", err
		}

		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cp -R '%s'/* '%s'", projectDir, srcDir))
		cmd.Dir, _ = os.Getwd()
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	return gothonDir, nil
}

func closeSession(gothonDir string) {
	if strings.ToLower(os.Getenv("GOTHON_KEEP_TEMP_DIR")) != "true" {
		_ = os.RemoveAll(gothonDir)
	}
}
