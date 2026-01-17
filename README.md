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

Once we have a list of commands we are happy with, we just need to
loop through and write the corresponding assembly commands to a file.
I'm planning to use the Flat Assembler (fasm) for this and support
`Linux x86_64` for starters. I might revisit and add WebAssembly or
maybe even `Darwin ARM64`.

## Tooling

I would like to add random tooling to help with development. For example
an IR Dump where you can see the generated IR list:

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

This would mean that the compiler will need some optimisation flags,
these can be set to remove optimisations to ensure we are getting the
right IR list for a given bit of code.

## Intermediate Representation (IR)

```
SHIFT k
    Shift the data pointer by k cells.
    Equivalent to: dp += k.

ADD k
    Add k to the current cell value (*dp), wrapping mod 256.

ZERO
    Set the current cell value (*dp) to zero.

IN
    Read one byte from input and store it in the current cell (*dp).

OUT
    Write the current cell value (*dp) to output as a single byte.

JZ target
    Jump to instruction index target if the current cell value (*dp) is
    zero.

JNZ target
    Jump to instruction index target if the current cell value (*dp) is
    non-zero.
```

- `k` is a signed integer
- `target` is an instruction index in the IR stream
- All arithmetic on cell values is performed modulo 256



[Brainfuck]: https://en.wikipedia.org/wiki/Brainfuck
