package util

import (
	"fmt"
	"os"
)

func Error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format + "\n", a...)
	os.Exit(1)
}
