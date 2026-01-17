# Brainfuck Compiler (bfcc)

default:
    @just --list

# Build the compiler
build:
    @mkdir -p ./bin
    go build -o bin/bfcc ./cmd/bfcc

# Run a brainfuck file (with -O2 optimization by default)
run file:
    go run ./cmd/bfcc run {{file}}

# Dump tokenizer output
tokens file:
    go run ./cmd/bfcc tokens {{file}}

# Dump IR (default -O 0, or specify -O 1/-O 2)
ir file *opts:
    go run ./cmd/bfcc ir {{opts}} {{file}}

# Output GAS assembly (x86_64 Linux)
asm file *opts:
    go run ./cmd/bfcc asm {{opts}} {{file}}

# Compile to ELF64 executable (x86_64 Linux)
compile file *opts:
    go run ./cmd/bfcc build {{opts}} {{file}}
