package memory

import (
	"errors"
	"io"
	"net"
	"sync"
	"tonysoft.com/gothon/pkg/log"
)

const (
	boolLength = 1
)

type BoolRegister struct {
	RegisterBase
	mut sync.Mutex
	val bool
}

func (r *BoolRegister) Init() {
	for i, s := range r.settersIn {
		go r.processSetter(s, r.settersOut[i])
	}

	for i, g := range r.gettersIn {
		go r.processGetter(g, r.gettersOut[i])
	}
}

func (r *BoolRegister) processSetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, boolLength)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			r.mut.Lock()
			r.val = inBytes[0] != 0
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:bool:set:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:bool:set:read:error: %v", readErr)
			return
		}
	}
}

func (r *BoolRegister) processGetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, boolLength)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				if r.val {
					outBytes[0] = 1
				} else {
					outBytes[0] = 0
				}
				r.mut.Unlock()

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:bool:get:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:bool:get:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:bool:get:read:error: %v", readErr)
			return
		}
	}
}
