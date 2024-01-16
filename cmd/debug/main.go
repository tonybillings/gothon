package main

import (
	"context"
	"os"
	"time"
	"tonysoft.com/gothon/pkg/console"
	"tonysoft.com/gothon/pkg/ipc"
	"tonysoft.com/gothon/pkg/log"
)

func main() {
	err := os.Setenv("GOTHON_KEEP_TEMP_DIR", "true")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	err = ipc.Start(ctx, cancel, "/projects/dev/go/gothon/test/queue", 5, "test1")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	intRecv := false
	go func() {
		<-ctx.Done()
		if intRecv {
			return
		}

		log.StopTime()
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()

	console.WaitForInterrupt()
	log.Warnf("Interrupt received at %s", time.Now().Format(log.TimeFormat))
	intRecv = true
	cancel()
	time.Sleep(3 * time.Second)
	os.Exit(130)
}
