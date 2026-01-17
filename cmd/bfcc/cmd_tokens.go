package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lcox74/bfcc/internal/core"
)

func cmdTokens(args []string) {
	fs := flag.NewFlagSet("tokens", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bfcc tokens <file>")
		os.Exit(1)
	}
	fs.Parse(args)

	if fs.NArg() != 1 {
		fs.Usage()
	}

	file := filepath.Clean(fs.Arg(0))
	src := readSource(file)

	tokens := core.Tokenize(src)
	for _, tok := range tokens {
		fmt.Printf("%d:%d\t%v\n", tok.Pos.Line, tok.Pos.Column, tok.Kind)
	}
}
