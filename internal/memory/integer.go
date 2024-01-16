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
	int64Length = 8
)

type IntRegister struct {
	RegisterBase
	mut sync.Mutex
	val int64
}

func (r *IntRegister) Init() {
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

	for i, m := range r.multipliersIn {
		go r.processMultiplier(m, r.multipliersOut[i])
	}

	for i, d := range r.dividersIn {
		go r.processDivider(d, r.dividersOut[i])
	}
}

func (r *IntRegister) processSetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, int64Length)
	var val int64
	var readErr, writeErr error
	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			val = int64(binary.BigEndian.Uint64(inBytes))
			r.mut.Lock()
			r.val = val
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:int:set:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:int:set:read:error: %v", readErr)
			return
		}
	}
}

func (r *IntRegister) processGetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, int64Length)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				binary.BigEndian.PutUint64(outBytes, uint64(r.val))
				r.mut.Unlock()

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:int:get:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:int:get:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:int:get:read:error: %v", readErr)
			return
		}
	}
}

func (r *IntRegister) processAdder(in io.Reader, out io.Writer) {
	inBytes := make([]byte, int64Length)
	var delta int64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			delta = int64(binary.BigEndian.Uint64(inBytes))
			r.mut.Lock()
			r.val += delta
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:int:add:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:int:add:read:error: %v", readErr)
			return
		}
	}
}

func (r *IntRegister) processSubtractor(in io.Reader, out io.Writer) {
	inBytes := make([]byte, int64Length)
	var delta int64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			delta = int64(binary.BigEndian.Uint64(inBytes))
			r.mut.Lock()
			r.val -= delta
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:int:sub:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:int:sub:read:error: %v", readErr)
			return
		}
	}
}

func (r *IntRegister) processMultiplier(in io.Reader, out io.Writer) {
	inBytes := make([]byte, int64Length)
	var multiplier int64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			multiplier = int64(binary.BigEndian.Uint64(inBytes))
			r.mut.Lock()
			r.val *= multiplier
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:int:mul:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:int:mul:read:error: %v", readErr)
			return
		}
	}
}

func (r *IntRegister) processDivider(in io.Reader, out io.Writer) {
	inBytes := make([]byte, int64Length)
	var divider int64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			divider = int64(binary.BigEndian.Uint64(inBytes))
			r.mut.Lock()
			r.val /= divider
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:int:div:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:int:div:read:error: %v", readErr)
			return
		}
	}
}
