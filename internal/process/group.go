package process

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
	"tonysoft.com/gothon/pkg/log"
)

type Group struct {
	startWaitGroup sync.WaitGroup
	stopWaitGroup  sync.WaitGroup
	stdoutChan     chan string
	stderrChan     chan string
	commands       []*exec.Cmd
}

func (g *Group) Start() *sync.WaitGroup {
	g.startWaitGroup.Done()
	return &g.stopWaitGroup
}

func (g *Group) Stop() {
	for _, cmd := range g.commands {
		if cmd.ProcessState != nil {
			if !cmd.ProcessState.Exited() {
				_ = cmd.Process.Signal(syscall.SIGINT)
			}
		}
	}
}

func (g *Group) StdOut() <-chan string {
	return g.stdoutChan
}

func (g *Group) StdErr() <-chan string {
	return g.stderrChan
}

func NewGroup(rootDir string, count int, args string) *Group {
	g := &Group{}
	g.stdoutChan = make(chan string, 1024)
	g.stderrChan = make(chan string, 1024)
	g.startWaitGroup.Add(1)
	g.stopWaitGroup.Add(count)

	for i := 0; i < count; i++ {
		cmd := exec.Command("python", "-u", "-m", args)
		cmd.Dir = filepath.Join(rootDir, strconv.Itoa(i))

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		go func(nodeId int) {
			scanner := bufio.NewScanner(stdout)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := scanner.Text()
				g.stdoutChan <- fmt.Sprintf("[%d] %s", nodeId, line)
			}
		}(i)

		go func(nodeId int) {
			scanner := bufio.NewScanner(stderr)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := scanner.Text()
				g.stderrChan <- fmt.Sprintf("[%d] %s", nodeId, line)
			}
		}(i)

		go func() {
			g.startWaitGroup.Wait()

			err := cmd.Start()
			if err != nil {
				log.Errorf("command start error: %v", err)
			}

			err = cmd.Wait()
			if err != nil && errors.Is(err, syscall.EINTR) {
				log.Errorf("command wait error: %v", err)
			}

			g.stopWaitGroup.Done()
		}()

		g.commands = append(g.commands, cmd)
	}

	go func() {
		g.stopWaitGroup.Wait()
		close(g.stdoutChan)
		close(g.stderrChan)
	}()

	time.Sleep(time.Second)

	return g
}
