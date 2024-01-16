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
	ctx, cancel := context.WithCancel(context.Background())

	nodeCount, nodeArgs, err := console.ParseArgs()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	err = ipc.Start(ctx, cancel, ".", nodeCount, nodeArgs)
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
