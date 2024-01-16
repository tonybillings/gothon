package io

import (
	"path/filepath"
	"strconv"
)

type SocketArray []*DomainSocket

func (a SocketArray) Listen() error {
	for _, socket := range a {
		err := socket.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a SocketArray) Close() {
	for _, socket := range a {
		socket.Close()
	}
}

func (a SocketArray) Get(path string) *DomainSocket {
	for _, socket := range a {
		if socket.path == path {
			return socket
		}
	}
	return nil
}

func NewSocketArray(basePath string, socketPaths []string, nodeCount int) SocketArray {
	socketArray := make([]*DomainSocket, 0)
	for i := 0; i < nodeCount; i++ {
		for _, path := range socketPaths {
			socketPath, _ := filepath.Abs(filepath.Join(basePath, strconv.Itoa(i), path))
			socket := NewDomainSocket(socketPath, path)
			socketArray = append(socketArray, socket)
		}
	}
	return socketArray
}
