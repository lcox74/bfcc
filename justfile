# Brainfuck Compiler (bfcc)

default:
    @just --list

# Build the compiler
build:
    go build -o bin/bfcc ./cmd/bfcc

# Run the compiler on a file
run file:
    go run ./cmd/bfcc {{file}}
