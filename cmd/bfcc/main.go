package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lcox74/bfcc/internal/core"
	"github.com/lcox74/bfcc/internal/vm"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage: bfcc <command> [options] <file>

commands:
  run [-O level] <file>    Run the program (default -O 2)
  tokens <file>            Dump tokenizer output
  ir [-O level] <file>     Dump IR (default -O 0)`)
	os.Exit(1)
}

func parseOptLevel(level int) core.OptLevel {
	switch level {
	case 0:
		return core.O0
	case 1:
		return core.O1
	case 2:
		return core.O2
	default:
		fmt.Fprintf(os.Stderr, "invalid optimization level: %d (must be 0, 1, or 2)\n", level)
		os.Exit(1)
	}
	return core.O0
}

func readSource(file string) []byte {
	src, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return src
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "tokens":
		cmdTokens(args)
	case "ir":
		cmdIR(args)
	case "run":
		cmdRun(args)
	default:
		usage()
	}
}

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

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	optLevel := fs.Int("O", 2, "optimization level (0, 1, or 2)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: bfcc run [-O level] <file>")
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

	interpreter := vm.NewVM()
	if err := interpreter.Run(ops); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
