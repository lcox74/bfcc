# Brainfuck Compiler (bfcc)

I'm trying to learn compilers again, did it during University briefly
in the COMP4403 course. We created a compiler in Java for the PL0
language. I say "we created", but it was more that we added features to
an existing implementation of the compiler created for the course.

So instead I'm going to build an **ahead-of-time (AOT)** compiler for
the infamous [Brainfuck] language. As mentioned, this is mainly for
learning purposes as no sane person would actively go out of their way
to write a rigid compiler for an esoteric programming language.

The language isn't new to me, I have built an interpreter for the 
language before. But I want to build a proper compiler which also
performs optimisations on the source code before generating some form
of executable.

## Process

### Tokenise

Brainfuck is only made from 8 characters:

- `>`: Shift Right
- `<`: Shift Left
- `+`: Increment
- `-`: Decrement
- `.`: Output
- `,`: Input
- `[`: Jump if Zero
- `]`: Jump if Not Zero

These first need to be tokenised so we can remove invalid characters
like comments. There isn't a dedicated comment indicator; any character
that isn't in the set is simply ignored. These characters can appear
throughout the code, so we need to tokenise to filter them out and
resolve address positions for jumps.

### IR Builder

Once we tokenise we need to convert tokens into operations also known
as an Intermediate Representation (IR). Because we want to add 
optimisations, we can add counters/values to the operations so:

```
++++ = ADD      +4
---- = ADD      -4
>>>> = SHIFT    +4
[    = JZ       <target>
```

### IR Optimiser

After this stage we can then perform more optimisations ontop like:

- Folding Adjacent Operations:
    - `ADD +4, ADD -4 = ADD 0`
- Removing No Operations:
    - `ADD 0`
    - `SHIFT 0`
    - `JZ <target>, JNZ <target>` (`[]`)
- Detect Zeroing Loops (`[-]`, `[+]`) and replace with `ZERO`

### Codegen

Once we have optimised IR, we generate GAS (GNU Assembler) output
targeting x86_64 Linux. The code generator:

- Uses R13 as the tape base pointer and R12 as the data pointer offset. The
  `r12-r15` are registers that are usually safe to use.
- Allocates a 30,000 byte tape in BSS (global variable)
- Emits syscalls for I/O (read/write) via helper functions
- Generates labels only where needed (jump targets)

To compile and run the generated assembly:

```bash
bfcc asm program.bf              # generates program.s
as -o program.o program.s        # assemble
ld -o program program.o          # link
./program                        # run
```

## Usage

```bash
bfcc <command> [options] <file>

commands:
  run [-O level] <file>          Run the program via VM (default -O 2)
  asm [-O level] [-o out] <file> Output GAS assembly (x86_64 Linux)
  tokens <file>                  Dump tokenizer output
  ir [-O level] <file>           Dump IR (default -O 0)
```

### IR Dump

The `ir` command dumps the intermediate representation:

```
000: ADD +6
001: JZ 008
002: SHIFT +1
003: ADD +10
004: SHIFT -1
005: ADD -1
006: JNZ 001
007: SHIFT +1
008: ADD +5
009: OUT
```

Use `-O 0`, `-O 1`, or `-O 2` to see IR at different optimisation levels.

## Documentation

- [Intermediate Representation (IR)](docs/ir.md)
- [IR to Assembly Mapping](docs/ir-to-asm.md)

[Brainfuck]: https://en.wikipedia.org/wiki/Brainfuck
