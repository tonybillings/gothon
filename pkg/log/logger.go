package log

import (
	"fmt"
	"os"
)

func Info(msg any) {
	_, e := fmt.Fprintf(os.Stdout, "\r[gothon] %v\n", msg)
	if e != nil {
		panic(e)
	}
}

func Infof(msg string, args ...any) {
	_, e := fmt.Fprintf(os.Stdout, "\r[gothon] %s\n", fmt.Sprintf(msg, args...))
	if e != nil {
		panic(e)
	}
}

func Warn(msg any) {
	_, e := fmt.Fprintf(os.Stdout, "\r\033[1;33m[gothon] %v\033[0m\n", msg)
	if e != nil {
		panic(e)
	}
}

func Warnf(msg string, args ...any) {
	_, e := fmt.Fprintf(os.Stdout, "\r\033[1;33m[gothon] %s\033[0m\n", fmt.Sprintf(msg, args...))
	if e != nil {
		panic(e)
	}
}

func Error(err any) {
	_, e := fmt.Fprintf(os.Stderr, "\r\033[1;31m[gothon] %v\033[0m\n", err)
	if e != nil {
		panic(e)
	}
}

func Errorf(err string, args ...any) {
	_, e := fmt.Fprintf(os.Stderr, "\r\033[1;31m[gothon] %s\033[0m\n", fmt.Sprintf(err, args...))
	if e != nil {
		panic(e)
	}
}
