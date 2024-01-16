package ipc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tonysoft.com/gothon/internal/process"
	"tonysoft.com/gothon/pkg/log"
)

func initProcessGroup(gothonDir string, nodeCount int, args string, cancel context.CancelFunc) *process.Group {
	pg := process.NewGroup(filepath.Join(gothonDir, "src"), nodeCount, args)

	go handleStdout(pg.StdOut())
	go handleStderr(pg.StdErr())

	time.Sleep(time.Second)

	wg := pg.Start()
	log.StartTime()

	go func() {
		wg.Wait()
		cancel()
	}()

	return pg
}

func handleStdout(stdout <-chan string) {
	for {
		select {
		case msg, ok := <-stdout:
			if !ok {
				return
			}
			_, e := fmt.Fprintf(os.Stdout, "\r%s\n", msg)
			if e != nil {
				return
			}
		}
	}
}

func handleStderr(stderr <-chan string) {
	for {
		select {
		case err, ok := <-stderr:
			if !ok {
				return
			}
			_, e := fmt.Fprintf(os.Stderr, "\r\033[31m%s\033[0m\n", err)
			if e != nil {
				return
			}
		}
	}
}
