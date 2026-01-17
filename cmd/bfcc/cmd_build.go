package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lcox74/bfcc/internal/codegen/linux"
	"github.com/lcox74/bfcc/internal/core"
)

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	optLevel := fs.Int("O", 2, "optimization level (0, 1, or 2)")
	output := fs.String("o", "", "output file (default: input file without extension)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bfcc build [-O level] [-o output] <file>")
		fmt.Fprintln(os.Stderr, "\nProduces a native ELF64 Linux executable directly.")
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
		outFile = strings.TrimSuffix(file, ".bf")
	}

	// Compile to IR
	tokens := core.Tokenize(src)
	ops, err := core.Lower(tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ops = core.OptimiseWithLevel(ops, level)

	// Generate ELF binary
	gen := linux.NewX86_64Generator(ops)
	binary := gen.GenerateELF()

	// Write executable file with executable permissions
	if err := os.WriteFile(outFile, binary, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("built %s -> %s\n", file, outFile)
}
