package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lcox74/bfcc/internal/core"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage: bfcc <command> [options] <file>

commands:
  run [-O level] <file>          Run the program (default -O 2)
  asm [-O level] [-o out] <file> Output GAS assembly (x86_64 Linux)
  tokens <file>                  Dump tokenizer output
  ir [-O level] <file>           Dump IR (default -O 0)`)
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
	file = filepath.Clean(file)
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
	case "asm":
		cmdAsm(args)
	default:
		usage()
	}
}
