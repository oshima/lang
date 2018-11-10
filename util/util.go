package util

import (
	"fmt"
	"os"
)

func Align(n int, boundary int) int {
	return (n + boundary - 1) & -boundary
}

func Error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
