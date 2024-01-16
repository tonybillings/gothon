package console

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func ParseArgs() (nodeCount int, nodeArgs string, err error) {
	if len(os.Args) < 2 {
		return -1, "", errors.New("missing node count from 'gothon' command")
	}

	if len(os.Args) < 3 {
		return -1, "", errors.New("missing module name from 'gothon' command")
	}

	nodeCount, err = strconv.Atoi(os.Args[1])
	if err != nil {
		return -1, "", fmt.Errorf("failed to parse 'node count' argument: %w", err)
	}

	if len(os.Args) > 2 {
		nodeArgs = strings.Join(os.Args[2:], " ")
	}

	return nodeCount, nodeArgs, nil
}

func WaitForInterrupt() {
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	<-intChan
}
