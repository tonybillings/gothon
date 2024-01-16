package config

import "sync/atomic"

const (
	DefaultStringRegisterBufferSize uint32 = 65536
)

var (
	stringRegisterBufferSize = atomic.Uint32{}
)

func GetStringRegisterBufferSize() uint32 {
	configuredSize := stringRegisterBufferSize.Load()
	if configuredSize != 0 {
		return configuredSize
	}
	return DefaultStringRegisterBufferSize
}

func SetStringRegisterBufferSize(value uint32) {
	stringRegisterBufferSize.Store(value)
}
