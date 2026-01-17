package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lcox74/bfcc/internal/codegen/gas"
	"github.com/lcox74/bfcc/internal/core"
)

func cmdAsm(args []string) {
	fs := flag.NewFlagSet("asm", flag.ExitOnError)
	optLevel := fs.Int("O", 2, "optimization level (0, 1, or 2)")
	output := fs.String("o", "", "output file (default: input file with .s extension)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bfcc asm [-O level] [-o output] <file>")
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

	// Determine output filename
	outFile := *output
	if outFile == "" {
		outFile = strings.TrimSuffix(file, ".bf") + ".s"
	}

	// Compile to IR
	tokens := core.Tokenize(src)
	ops, err := core.Lower(tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ops = core.OptimiseWithLevel(ops, level)

	// Generate assembly
	gen := gas.NewGenerator(ops)
	asm := gen.Generate()

	// Write assembly file
	if err := os.WriteFile(outFile, []byte(asm), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("generated %s -> %s\n", file, outFile)
}
