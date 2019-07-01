package main

import (
	"io/ioutil"
	"os"

	"github.com/oshima/lang/gen"
	"github.com/oshima/lang/parse"
	"github.com/oshima/lang/scan"
	"github.com/oshima/lang/sema"
)

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	runes := []rune(string(bytes))
	tokens := scan.Scan(runes)
	prog := parse.Parse(tokens)
	sema.Analyze(prog)
	gen.Generate(prog)
}
