package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lcox74/bfcc/internal/core"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bfcc <file>")
		os.Exit(1)
	}

	file := filepath.Clean(os.Args[1])

	src, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Tokenizer:")
	tokens := core.Tokenize(src)
	for _, tok := range tokens {
		fmt.Printf("%d:%d\t%v\n", tok.Pos.Line, tok.Pos.Column, tok.Kind)
	}

	fmt.Println("IR:")
	ops, err := core.Lower(tokens)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dump := core.Dump(ops)
	fmt.Println(dump)

	fmt.Println("IR (optimise):")
	ops = core.Optimise(ops)
	dump = core.Dump(ops)
	fmt.Println(dump)
}
