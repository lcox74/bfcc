# IR to Assembly Mapping

This document describes how BFCC IR operations are lowered to x86_64 GAS 
(AT&T syntax) assembly for Linux.

## Register Allocation

| Register | Purpose |
|----------|---------|
| `%r13` | Tape base address (constant after init) |
| `%r12` | Data pointer offset (cell index) |
| `%rax` | Syscall number |
| `%rdi` | Syscall arg 1 (fd) |
| `%rsi` | Syscall arg 2 (buffer address) |
| `%rdx` | Syscall arg 3 (count) |

Current cell address is computed as `(%r13,%r12)` - base plus offset 
addressing.

## IR Operations

### SHIFT k

Move data pointer by k cells.

```asm
# SHIFT +3
addq $3, %r12

# SHIFT -2
subq $2, %r12
```

### ADD k

Add k to current cell (wraps at 256).

```asm
# ADD +5
addb $5, (%r13,%r12)

# ADD -1
subb $1, (%r13,%r12)
```

### ZERO

Clear current cell. Optimized form of `[-]` loops.

```asm
movb $0, (%r13,%r12)
```

### IN

Read one byte from stdin into current cell. Calls helper function.

```asm
call _bf_read
```

### OUT

Write current cell to stdout. Calls helper function.

```asm
call _bf_write
```

### JZ target

Jump to target if current cell is zero. Opens a loop.

```asm
testb $0xff, (%r13,%r12)
jz .jt_target
```

### JNZ target

Jump to target if current cell is non-zero. Closes a loop.

```asm
testb $0xff, (%r13,%r12)
jnz .jt_target
```

## Helper Functions

I/O operations use helper functions to reduce code size. Emitted once at end
of program.

```asm
_bf_read:
    leaq (%r13,%r12), %rsi
    xorq %rax, %rax
    xorq %rdi, %rdi
    movq $1, %rdx
    syscall
    ret

_bf_write:
    leaq (%r13,%r12), %rsi
    movq $1, %rax
    movq $1, %rdi
    movq $1, %rdx
    syscall
    ret
```

## Syscall Numbers

| Syscall | Number | Signature |
|---------|--------|-----------|
| read | 0 | `read(fd, buf, count)` |
| write | 1 | `write(fd, buf, count)` |
| exit | 60 | `exit(code)` |

## Program Structure

```asm
.section .bss
    .lcomm tape, 30000       # 30k cell tape

.section .text
.globl _start
_start:
    movq $tape, %r13         # tape base
    xorq %r12, %r12          # dp = 0

    # ... IR operations ...

.jt_5:                       # only jump targets get labels
    # ... more operations ...

    movq $60, %rax           # exit(0)
    xorq %rdi, %rdi
    syscall

_bf_read:
    # ... helper impl ...

_bf_write:
    # ... helper impl ...
```

Labels are only emitted for jump targets (indices referenced by `JZ`/`JNZ`
instructions). Helper functions are emitted after the exit syscall.

## Resources

- [GAS Examples](https://cs.lmu.edu/~ray/notes/gasexamples/) - GNU Assembler syntax and examples
- [Linux System Call Table for x86_64](https://blog.rchapman.org/posts/Linux_System_Call_Table_for_x86_64/) - Syscall numbers and register conventions
