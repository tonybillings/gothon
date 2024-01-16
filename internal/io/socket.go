package io

import (
	"net"
	"os"
	"strings"
)

type DomainSocket struct {
	Tag string

	path    string
	connIn  net.PacketConn
	connOut *net.UnixConn
	dest    *net.UnixAddr
}

func (s *DomainSocket) Init() error {
	err := s.createBaseDirectory()
	if err != nil {
		return err
	}

	addr, err := net.ResolveUnixAddr("unixgram", s.path)
	if err != nil {
		return err
	}
	s.dest = addr

	if strings.HasSuffix(s.path, "_out") || strings.HasSuffix(s.path, "_ok") {
		return nil
	}

	conn, err := net.ListenPacket("unixgram", s.path)
	if err != nil {
		return err
	}
	s.connIn = conn

	return nil
}

func (s *DomainSocket) Read(buffer []byte) (int, error) {
	n, _, err := s.connIn.ReadFrom(buffer)
	return n, err
}

func (s *DomainSocket) Write(data []byte) (int, error) {
	if s.connOut == nil {
		var err error
		s.connOut, err = net.DialUnix("unixgram", nil, s.dest)
		if err != nil {
			return -1, err
		}
	}
	return s.connOut.Write(data)
}

func (s *DomainSocket) Close() {
	if s.connIn != nil {
		_ = s.connIn.Close()
	}
	if s.connOut != nil {
		_ = s.connOut.Close()
	}
}

func (s *DomainSocket) Path() string {
	return s.path
}

func (s *DomainSocket) createBaseDirectory() error {
	pathParts := strings.Split(s.path, "/")
	dirPath := strings.TrimSuffix(s.path, "/"+pathParts[len(pathParts)-1])
	return os.MkdirAll(dirPath, 0775)
}

func NewDomainSocket(path string, tags ...string) *DomainSocket {
	s := &DomainSocket{
		path: path,
	}

	for _, t := range tags {
		s.Tag += t + " "
	}
	s.Tag = strings.TrimSpace(s.Tag)

	return s
}
