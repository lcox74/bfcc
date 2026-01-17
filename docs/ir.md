# Intermediate Representation (IR)

The compiler uses a simple intermediate representation with seven operations.

## Operations

### SHIFT k

Shift the data pointer by k cells.

```
Equivalent to: dp += k
```

### ADD k

Add k to the current cell value (*dp), wrapping mod 256.

```
Equivalent to: *dp = (*dp + k) % 256
```

### ZERO

Set the current cell value (*dp) to zero. This is an optimised replacement
for the common `[-]` and `[+]` patterns.

```
Equivalent to: *dp = 0
```

### IN

Read one byte from input and store it in the current cell (*dp).

```
Equivalent to: *dp = getchar()
```

### OUT

Write the current cell value (*dp) to output as a single byte.

```
Equivalent to: putchar(*dp)
```

### JZ target

Jump to instruction index `target` if the current cell value (*dp) is zero.

```
Equivalent to: if (*dp == 0) goto target
```

### JNZ target

Jump to instruction index `target` if the current cell value (*dp) is
non-zero.

```
Equivalent to: if (*dp != 0) goto target
```

## Notes

- `k` is a signed integer
- `target` is an instruction index in the IR stream
- All arithmetic on cell values is performed modulo 256

## Example

Given the Brainfuck code `++++++[>++++++++++<-]>+++++.`, the IR at `-O 2`
would be:

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

This prints the character 'A' (ASCII 65 = 6 * 10 + 5).
