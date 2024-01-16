package memory

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"tonysoft.com/gothon/internal/memory/config"
	"tonysoft.com/gothon/pkg/log"
)

type StringRegister struct {
	RegisterBase
	mut        sync.Mutex
	val        string
	bufferSize uint32
}

func (r *StringRegister) Init() {
	for i, s := range r.settersIn {
		go r.processSetter(s, r.settersOut[i])
	}

	for i, g := range r.gettersIn {
		go r.processGetter(g, r.gettersOut[i])
	}

	for i, a := range r.addersIn {
		go r.processAdder(a, r.addersOut[i])
	}

	for i, s := range r.subtractorsIn {
		go r.processSubtractor(s, r.subtractorsOut[i])
	}

	r.bufferSize = config.GetStringRegisterBufferSize()
}

func (r *StringRegister) processSetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, float64Length)
	count := 0
	var readErr, writeErr error

	for {
		count, readErr = in.Read(inBytes)
		if readErr == nil {
			r.mut.Lock()
			r.val = string(inBytes[:count])
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:string:set:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:string:set:read:error: %v", readErr)
			return
		}
	}
}

func (r *StringRegister) processGetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	var outBytes []byte
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				outBytes = []byte(r.val)
				r.mut.Unlock()

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:string:get:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:string:get:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:string:get:read:error: %v", readErr)
			return
		}
	}
}

func (r *StringRegister) processAdder(in io.Reader, out io.Writer) {
	inBytes := make([]byte, r.bufferSize)
	count := 0
	var readErr, writeErr error

	for {
		count, readErr = in.Read(inBytes)
		if readErr == nil {
			r.mut.Lock()
			r.val += string(inBytes[:count])
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:string:add:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:string:add:read:error: %v", readErr)
			return
		}
	}
}

func (r *StringRegister) processSubtractor(in io.Reader, out io.Writer) {
	inBytes := make([]byte, r.bufferSize)
	count := 0
	var readErr, writeErr error

	for {
		count, readErr = in.Read(inBytes)
		if readErr == nil {
			r.mut.Lock()
			r.val = strings.TrimSuffix(r.val, string(inBytes[:count]))
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:string:sub:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:string:sub:read:error: %v", readErr)
			return
		}
	}
}
