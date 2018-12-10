package main

import (
	"flag"
	"github.com/k0kubun/pp"
	"github.com/oshjma/lang/gen"
	"github.com/oshjma/lang/parse"
	"github.com/oshjma/lang/scan"
	"github.com/oshjma/lang/sema"
	"github.com/oshjma/lang/util"
	"io/ioutil"
	"os"
)

func main() {
	debug := flag.Bool("d", false, "print tokens and AST for debug")
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		util.Error("Failed to read source code from stdin")
	}

	runes := []rune(string(bytes))

	tokens := scan.Scan(runes)
	if *debug {
		pp.Fprintln(os.Stderr, tokens)
	}

	prog := parse.Parse(tokens)
	if *debug {
		pp.Fprintln(os.Stderr, prog)
	}

	meta := sema.Analyze(prog)

	gen.Generate(prog, meta)
}
