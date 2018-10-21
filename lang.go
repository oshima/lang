package main

import (
	"flag"
	"io/ioutil"
	"os"
	"github.com/k0kubun/pp"
	"github.com/oshjma/lang/gen"
	"github.com/oshjma/lang/parse"
	"github.com/oshjma/lang/scan"
	"github.com/oshjma/lang/util"
)

func main() {
	debug := flag.Bool("d", false, "print tokens and AST for debug")
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		util.Error("Failed to read source code from stdin")
	}

	tokens := scan.Scan(string(bytes))
	if *debug {
		pp.Fprintln(os.Stderr, tokens)
	}

	program := parse.Parse(tokens)
	if *debug {
		pp.Fprintln(os.Stderr, program)
	}

	if !*debug {
		gen.Generate(program)
	}
}
