package memory

import (
	"errors"
	"io"
	"net"
	"sync"
	"tonysoft.com/gothon/pkg/log"
)

type MutexRegister struct {
	RegisterBase
	val sync.Mutex
}

func (r *MutexRegister) Init() {
	for i, l := range r.lockersIn {
		go r.processLocker(l, r.lockersOut[i])
	}

	for i, u := range r.unlockersIn {
		go r.processUnlocker(u, r.unlockersOut[i])
	}
}

func (r *MutexRegister) processLocker(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.val.Lock()

				_, writeErr = out.Write(syncBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:mutex:lock:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:mutex:lock:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:mutex:lock:read:error: %v", readErr)
			return
		}
	}
}

func (r *MutexRegister) processUnlocker(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.val.Unlock()

				_, writeErr = out.Write(syncBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:mutex:unlock:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:mutex:unlock:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:mutex:unlock:read:error: %v", readErr)
			return
		}
	}
}
