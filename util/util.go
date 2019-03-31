package util

import (
	"fmt"
	"os"
)

// Error prints the error message and stops compiling.
func Error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
