package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lcox74/bfcc/internal/core"
)

func cmdIR(args []string) {
	fs := flag.NewFlagSet("ir", flag.ExitOnError)
	optLevel := fs.Int("O", 0, "optimization level (0, 1, or 2)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bfcc ir [-O level] <file>")
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)

	if fs.NArg() != 1 {
		fs.Usage()
	}

	level := parseOptLevel(*optLevel)
	file := filepath.Clean(fs.Arg(0))
	src := readSource(file)

	tokens := core.Tokenize(src)
	ops, err := core.Lower(tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ops = core.OptimiseWithLevel(ops, level)
	fmt.Print(core.Dump(ops))
}
