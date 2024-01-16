package memory

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"tonysoft.com/gothon/pkg/log"
)

const (
	float64Length = 8
)

type FloatRegister struct {
	RegisterBase
	mut sync.Mutex
	val float64
}

func (r *FloatRegister) Init() {
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

func (r *FloatRegister) processSetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, float64Length)
	var bits uint64
	var val float64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			bits = binary.LittleEndian.Uint64(inBytes)
			val = math.Float64frombits(bits)
			r.mut.Lock()
			r.val = val
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:float:set:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:float:set:read:error: %v", readErr)
			return
		}
	}
}

func (r *FloatRegister) processGetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, float64Length)
	var bits uint64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				bits = math.Float64bits(r.val)
				r.mut.Unlock()
				binary.LittleEndian.PutUint64(outBytes, bits)

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:float:get:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:float:get:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:float:get:read:error: %v", readErr)
			return
		}
	}
}

func (r *FloatRegister) processAdder(in io.Reader, out io.Writer) {
	inBytes := make([]byte, float64Length)
	var bits uint64
	var delta float64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			bits = binary.LittleEndian.Uint64(inBytes)
			delta = math.Float64frombits(bits)
			r.mut.Lock()
			r.val += delta
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:float:add:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:float:add:read:error: %v", readErr)
			return
		}
	}
}

func (r *FloatRegister) processSubtractor(in io.Reader, out io.Writer) {
	inBytes := make([]byte, float64Length)
	var bits uint64
	var delta float64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			bits = binary.LittleEndian.Uint64(inBytes)
			delta = math.Float64frombits(bits)
			r.mut.Lock()
			r.val -= delta
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:float:sub:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:float:sub:read:error: %v", readErr)
			return
		}
	}
}

func (r *FloatRegister) processMultiplier(in io.Reader, out io.Writer) {
	inBytes := make([]byte, float64Length)
	var bits uint64
	var multiplier float64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			bits = binary.LittleEndian.Uint64(inBytes)
			multiplier = math.Float64frombits(bits)
			r.mut.Lock()
			r.val *= multiplier
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:float:mul:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:float:mul:read:error: %v", readErr)
			return
		}
	}
}

func (r *FloatRegister) processDivider(in io.Reader, out io.Writer) {
	inBytes := make([]byte, float64Length)
	var bits uint64
	var divider float64
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			bits = binary.LittleEndian.Uint64(inBytes)
			divider = math.Float64frombits(bits)
			r.mut.Lock()
			r.val /= divider
			r.mut.Unlock()

			_, writeErr = out.Write(syncBytes)
			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:float:div:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:float:div:read:error: %v", readErr)
			return
		}
	}
}
