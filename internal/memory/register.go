package memory

import (
	"io"
	"sync"
	"tonysoft.com/gothon/internal/queue"
)

var (
	nakByte  = byte(21)
	syncByte = byte(22)

	nakBytes  = []byte{nakByte}
	syncBytes = []byte{syncByte}
)

type QueueRegisterType interface {
	queue.Fifo[bool] | queue.Fifo[int64] | queue.Fifo[float64] | queue.Fifo[string] |
		queue.Lifo[bool] | queue.Lifo[int64] | queue.Lifo[float64] | queue.Lifo[string]
}

type RegisterType interface {
	bool | int64 | float64 | string |
		sync.Mutex | *sync.WaitGroup |
		QueueRegisterType
}

type Register interface {
	ID() string

	AddSetterIn(io.Reader)
	AddSetterOut(io.Writer)

	AddGetterIn(io.Reader)
	AddGetterOut(io.Writer)
	AddGetterOk(io.Writer)

	AddAdderIn(io.Reader)
	AddAdderOut(io.Writer)

	AddSubtractorIn(io.Reader)
	AddSubtractorOut(io.Writer)

	AddMultiplierIn(io.Reader)
	AddMultiplierOut(io.Writer)

	AddDividerIn(io.Reader)
	AddDividerOut(io.Writer)

	AddLockerIn(io.Reader)
	AddLockerOut(io.Writer)

	AddUnlockerIn(io.Reader)
	AddUnlockerOut(io.Writer)

	AddSizeCallerIn(io.Reader)
	AddSizeCallerOut(io.Writer)

	AddEmptyCallerIn(io.Reader)
	AddEmptyCallerOut(io.Writer)

	AddFullCallerIn(io.Reader)
	AddFullCallerOut(io.Writer)

	Init()
}

type RegisterBase struct {
	id string

	settersIn  []io.Reader
	settersOut []io.Writer

	gettersIn  []io.Reader
	gettersOut []io.Writer
	gettersOk  []io.Writer

	addersIn  []io.Reader
	addersOut []io.Writer

	subtractorsIn  []io.Reader
	subtractorsOut []io.Writer

	multipliersIn  []io.Reader
	multipliersOut []io.Writer

	dividersIn  []io.Reader
	dividersOut []io.Writer

	lockersIn  []io.Reader
	lockersOut []io.Writer

	unlockersIn  []io.Reader
	unlockersOut []io.Writer

	sizeCallersIn  []io.Reader
	sizeCallersOut []io.Writer

	emptyCallersIn  []io.Reader
	emptyCallersOut []io.Writer

	fullCallersIn  []io.Reader
	fullCallersOut []io.Writer
}

func (r *RegisterBase) ID() string {
	return r.id
}

func (r *RegisterBase) AddSetterIn(reader io.Reader) {
	r.settersIn = append(r.settersIn, reader)
}

func (r *RegisterBase) AddSetterOut(writer io.Writer) {
	r.settersOut = append(r.settersOut, writer)
}

func (r *RegisterBase) AddGetterIn(reader io.Reader) {
	r.gettersIn = append(r.gettersIn, reader)
}

func (r *RegisterBase) AddGetterOut(writer io.Writer) {
	r.gettersOut = append(r.gettersOut, writer)
}

func (r *RegisterBase) AddGetterOk(writer io.Writer) {
	r.gettersOk = append(r.gettersOk, writer)
}

func (r *RegisterBase) AddAdderIn(reader io.Reader) {
	r.addersIn = append(r.addersIn, reader)
}

func (r *RegisterBase) AddAdderOut(writer io.Writer) {
	r.addersOut = append(r.addersOut, writer)
}

func (r *RegisterBase) AddSubtractorIn(reader io.Reader) {
	r.subtractorsIn = append(r.subtractorsIn, reader)
}

func (r *RegisterBase) AddSubtractorOut(writer io.Writer) {
	r.subtractorsOut = append(r.subtractorsOut, writer)
}

func (r *RegisterBase) AddMultiplierIn(reader io.Reader) {
	r.multipliersIn = append(r.multipliersIn, reader)
}

func (r *RegisterBase) AddMultiplierOut(writer io.Writer) {
	r.multipliersOut = append(r.multipliersOut, writer)
}

func (r *RegisterBase) AddDividerIn(reader io.Reader) {
	r.dividersIn = append(r.dividersIn, reader)
}

func (r *RegisterBase) AddDividerOut(writer io.Writer) {
	r.dividersOut = append(r.dividersOut, writer)
}

func (r *RegisterBase) AddLockerIn(reader io.Reader) {
	r.lockersIn = append(r.lockersIn, reader)
}

func (r *RegisterBase) AddLockerOut(writer io.Writer) {
	r.lockersOut = append(r.lockersOut, writer)
}

func (r *RegisterBase) AddUnlockerIn(reader io.Reader) {
	r.unlockersIn = append(r.unlockersIn, reader)
}

func (r *RegisterBase) AddUnlockerOut(writer io.Writer) {
	r.unlockersOut = append(r.unlockersOut, writer)
}

func (r *RegisterBase) AddSizeCallerIn(reader io.Reader) {
	r.sizeCallersIn = append(r.sizeCallersIn, reader)
}

func (r *RegisterBase) AddSizeCallerOut(writer io.Writer) {
	r.sizeCallersOut = append(r.sizeCallersOut, writer)
}

func (r *RegisterBase) AddEmptyCallerIn(reader io.Reader) {
	r.emptyCallersIn = append(r.emptyCallersIn, reader)
}

func (r *RegisterBase) AddEmptyCallerOut(writer io.Writer) {
	r.emptyCallersOut = append(r.emptyCallersOut, writer)
}

func (r *RegisterBase) AddFullCallerIn(reader io.Reader) {
	r.fullCallersIn = append(r.fullCallersIn, reader)
}

func (r *RegisterBase) AddFullCallerOut(writer io.Writer) {
	r.fullCallersOut = append(r.fullCallersOut, writer)
}

func NewRegister[T RegisterType](id string, defaultValue any) Register {
	switch any(*new(T)).(type) {
	case bool:
		reg := &BoolRegister{}
		reg.id = id
		reg.val = defaultValue.(bool)
		return reg
	case int64:
		reg := &IntRegister{}
		reg.id = id
		reg.val = defaultValue.(int64)
		return reg
	case float64:
		reg := &FloatRegister{}
		reg.id = id
		reg.val = defaultValue.(float64)
		return reg
	case string:
		reg := &StringRegister{}
		reg.id = id
		reg.val = defaultValue.(string)
		return reg
	case sync.Mutex:
		reg := &MutexRegister{}
		reg.id = id
		return reg
	case *sync.WaitGroup:
		wg := defaultValue.(*sync.WaitGroup)
		reg := &WaitGroupRegister{}
		reg.id = id
		reg.val = wg
		return reg
	case queue.Fifo[bool]:
		q := queue.New[bool](uint64(defaultValue.(int64)))
		reg := &QueueRegister[bool]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Fifo[int64]:
		q := queue.New[int64](uint64(defaultValue.(int64)))
		reg := &QueueRegister[int64]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Fifo[float64]:
		q := queue.New[float64](uint64(defaultValue.(int64)))
		reg := &QueueRegister[float64]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Fifo[string]:
		q := queue.New[string](uint64(defaultValue.(int64)))
		reg := &QueueRegister[string]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Lifo[bool]:
		q := queue.New[bool](uint64(defaultValue.(int64)), true)
		reg := &QueueRegister[bool]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Lifo[int64]:
		q := queue.New[int64](uint64(defaultValue.(int64)), true)
		reg := &QueueRegister[int64]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Lifo[float64]:
		q := queue.New[float64](uint64(defaultValue.(int64)), true)
		reg := &QueueRegister[float64]{}
		reg.id = id
		reg.val = q
		return reg
	case queue.Lifo[string]:
		q := queue.New[string](uint64(defaultValue.(int64)), true)
		reg := &QueueRegister[string]{}
		reg.id = id
		reg.val = q
		return reg
	}

	return nil
}
