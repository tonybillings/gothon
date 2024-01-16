package memory

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"tonysoft.com/gothon/internal/memory/config"
	"tonysoft.com/gothon/internal/queue"
	"tonysoft.com/gothon/pkg/log"
)

type QueueRegister[T queue.ItemType] struct {
	RegisterBase
	mut        sync.Mutex
	val        queue.Queue[T]
	bufferSize uint32
	readVal    func(buff []byte, count int) T
	writeVal   func(val T, buff []byte) []byte
}

func (r *QueueRegister[T]) Init() {
	for i, s := range r.settersIn {
		go r.processSetter(s, r.settersOut[i])
	}

	for i, g := range r.gettersIn {
		go r.processGetter(g, r.gettersOut[i], r.gettersOk[i])
	}

	for i, c := range r.sizeCallersIn {
		go r.processSizeCaller(c, r.sizeCallersOut[i])
	}

	for i, c := range r.emptyCallersIn {
		go r.processEmptyCaller(c, r.emptyCallersOut[i])
	}

	for i, c := range r.fullCallersIn {
		go r.processFullCaller(c, r.fullCallersOut[i])
	}

	r.setBufferSize()
	r.setReadValFunc()
	r.setWriteValFunc()
}

func (r *QueueRegister[T]) setBufferSize() {
	switch any(*new(T)).(type) {
	case bool:
		r.bufferSize = boolLength
	case int64:
		r.bufferSize = int64Length
	case float64:
		r.bufferSize = float64Length
	case string:
		r.bufferSize = config.GetStringRegisterBufferSize()
	}
}

func (r *QueueRegister[T]) setReadValFunc() {
	switch any(*new(T)).(type) {
	case bool:
		r.readVal = func(buff []byte, count int) T {
			return any(buff[0] != 0).(T)
		}
	case int64:
		r.readVal = func(buff []byte, count int) T {
			return any(int64(binary.BigEndian.Uint64(buff))).(T)
		}
	case float64:
		r.readVal = func(buff []byte, count int) T {
			bits := binary.LittleEndian.Uint64(buff)
			return any(math.Float64frombits(bits)).(T)
		}
	case string:
		r.readVal = func(buff []byte, count int) T {
			return any(string(buff[:count])).(T)
		}
	}
}

func (r *QueueRegister[T]) setWriteValFunc() {
	switch any(*new(T)).(type) {
	case bool:
		r.writeVal = func(val T, buff []byte) []byte {
			if any(val).(bool) {
				buff[0] = 1
			} else {
				buff[0] = 0
			}
			return buff
		}
	case int64:
		r.writeVal = func(val T, buff []byte) []byte {
			binary.BigEndian.PutUint64(buff, uint64(any(val).(int64)))
			return buff
		}
	case float64:
		r.writeVal = func(val T, buff []byte) []byte {
			bits := math.Float64bits(any(val).(float64))
			binary.LittleEndian.PutUint64(buff, bits)
			return buff
		}
	case string:
		r.writeVal = func(val T, buff []byte) []byte {
			return []byte(any(val).(string))
		}
	}
}

func (r *QueueRegister[T]) processSetter(in io.Reader, out io.Writer) {
	inBytes := make([]byte, r.bufferSize)
	var val T
	var count int
	var readErr, writeErr error
	var ok bool

	for {
		count, readErr = in.Read(inBytes)
		if readErr == nil {
			val = r.readVal(inBytes, count)
			r.mut.Lock()
			ok = r.val.Put(val)
			r.mut.Unlock()

			if ok {
				_, writeErr = out.Write(syncBytes)
			} else {
				_, writeErr = out.Write(nakBytes)
			}

			if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
				log.Errorf("register:queue:set:write:error: %v", writeErr)
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:queue:set:read:error: %v", readErr)
			return
		}
	}
}

func (r *QueueRegister[T]) processGetter(in io.Reader, out io.Writer, ok io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, r.bufferSize)
	var readErr, writeErr error
	var val T
	var isNotEmpty bool

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				val, isNotEmpty = r.val.Get()
				r.mut.Unlock()

				if isNotEmpty {
					_, writeErr = ok.Write(syncBytes)
					if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
						log.Errorf("register:queue:get:write:error: %v", writeErr)
						return
					}

					outBytes = r.writeVal(val, outBytes)
					_, writeErr = out.Write(outBytes)
				} else {
					_, writeErr = out.Write(nakBytes)
				}

				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:queue:get:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:queue:get:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:queue:get:read:error: %v", readErr)
			return
		}
	}
}

func (r *QueueRegister[T]) processSizeCaller(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, int64Length)
	var readErr, writeErr error
	var size uint64

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				size = r.val.Size()
				r.mut.Unlock()
				binary.BigEndian.PutUint64(outBytes, size)

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:queue:size:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:queue:size:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:queue:size:read:error: %v", readErr)
			return
		}
	}
}

func (r *QueueRegister[T]) processEmptyCaller(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, boolLength)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				if r.val.Empty() {
					outBytes[0] = 1
				} else {
					outBytes[0] = 0
				}
				r.mut.Unlock()

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:queue:size:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:queue:size:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:queue:size:read:error: %v", readErr)
			return
		}
	}
}

func (r *QueueRegister[T]) processFullCaller(in io.Reader, out io.Writer) {
	inBytes := make([]byte, 1)
	outBytes := make([]byte, boolLength)
	var readErr, writeErr error

	for {
		_, readErr = in.Read(inBytes)
		if readErr == nil {
			if inBytes[0] == syncByte {
				r.mut.Lock()
				if r.val.Full() {
					outBytes[0] = 1
				} else {
					outBytes[0] = 0
				}
				r.mut.Unlock()

				_, writeErr = out.Write(outBytes)
				if writeErr != nil && !errors.Is(writeErr, net.ErrClosed) {
					log.Errorf("register:queue:size:write:error: %v", writeErr)
					return
				}
			} else {
				log.Errorf("register:queue:size:read:error: expected byte 22, got %d", inBytes[0])
				return
			}
		} else if !errors.Is(readErr, net.ErrClosed) {
			log.Errorf("register:queue:size:read:error: %v", readErr)
			return
		}
	}
}
