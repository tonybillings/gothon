package test

import "testing"

const (
	defaultNodeCount = 5
)

func TestBool(t *testing.T) {
	installGothon(t)
	runGothon(t, "bool", defaultNodeCount)
}

func TestInt(t *testing.T) {
	installGothon(t)
	runGothon(t, "int", defaultNodeCount)
}

func TestFloat(t *testing.T) {
	installGothon(t)
	runGothon(t, "float", defaultNodeCount)
}

func TestString(t *testing.T) {
	installGothon(t)
	runGothon(t, "str", defaultNodeCount)
}

func TestMutex(t *testing.T) {
	installGothon(t)
	runGothon(t, "mutex", defaultNodeCount)
}

func TestSync(t *testing.T) {
	installGothon(t)
	runGothon(t, "sync", defaultNodeCount)
}

func TestQueue(t *testing.T) {
	installGothon(t)
	runGothon(t, "queue", defaultNodeCount)
}
