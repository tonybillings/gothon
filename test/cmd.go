package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func installGothon(t *testing.T) {
	var err error
	cmd := exec.Command("./install.sh")
	cmd.Dir, err = filepath.Abs("..")
	if err != nil {
		t.Error(err)
		return
	}

	err = cmd.Run()
	if err != nil {
		t.Error(err)
		return
	}
}

func runGothon(t *testing.T, projectDir string, nodeCount int, modules ...string) {
	var err error

	run := func(cmd *exec.Cmd) {
		stdout, e := cmd.StdoutPipe()
		if e != nil {
			t.Error(e)
			return
		}

		stderr, e := cmd.StderrPipe()
		if e != nil {
			t.Error(e)
			return
		}

		go func() {
			buffer := make([]byte, 65536)
			for {
				count, readErr := stdout.Read(buffer)
				if readErr != nil {
					return
				}
				t.Logf("%s", strings.ReplaceAll(string(buffer[:count]), "\r", ""))
			}
		}()

		go func() {
			buffer := make([]byte, 65536)
			for {
				count, readErr := stderr.Read(buffer)
				if readErr != nil {
					return
				}
				t.Errorf("%s", strings.ReplaceAll(string(buffer[:count]), "\r", ""))
			}
		}()

		e = cmd.Start()
		if e != nil {
			t.Error(e)
			return
		}

		e = cmd.Wait()
		if e != nil {
			t.Error(e)
			return
		}

		time.Sleep(2 * time.Second)
	}

	if len(modules) == 0 {
		dir, e := filepath.Abs(projectDir)
		if e != nil {
			t.Error(e)
			return
		}

		files, e := os.ReadDir(dir)
		if e != nil {
			t.Error(e)
			return
		}

		for _, module := range files {
			cmd := exec.Command("gothon", strconv.Itoa(nodeCount), strings.TrimSuffix(module.Name(), ".py"))
			cmd.Dir = dir
			run(cmd)
		}
	} else {
		for _, module := range modules {
			cmd := exec.Command("gothon", strconv.Itoa(nodeCount), module)
			cmd.Dir, err = filepath.Abs(projectDir)
			if err != nil {
				t.Error(err)
				return
			}
			run(cmd)
		}
	}
}
