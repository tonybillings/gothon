package memory

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"tonysoft.com/gothon/pkg/log"
)

const (
	int32Length = 4
)

type WaitGroupRegister struct {
	RegisterBase
	val *sync.WaitGroup
}

func (r *WaitGroupRegister) Init() {
	for i, s := range r.settersIn {
		go r.processSetter(s, r.settersOut[i])
	}
}

func (r *WaitGroupRegister) processSetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, int32Length)
	var val int32
	var readErr, writeErr error
	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			val = int32(binary.BigEndian.Uint32(inBytes))
			switch {
			case val == 0:
				r.val.Wait()
			case val > 0:
				for i := int32(0); i < val; i++ {
					r.val.Done()
				}
			default:
				log.Error("register:sync:error: sync val must not be negative")
				return
			}

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:sync:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:sync:read:error: %v", readErr)
			return
		}
	}
}
