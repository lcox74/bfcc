// Package core provides the fundamental types and functions for the Brainfuck compiler.
//
// This package includes:
//   - Tokenizer: converts Brainfuck source text into a stream of tokens
//   - IR: intermediate representation for backend-agnostic code generation
//
// Brainfuck has eight commands, each represented by a single character:
//   - > : increment the data pointer
//   - < : decrement the data pointer
//   - + : increment the byte at the data pointer
//   - - : decrement the byte at the data pointer
//   - . : output the byte at the data pointer
//   - , : input a byte and store it at the data pointer
//   - [ : jump forward past matching ] if byte at pointer is zero
//   - ] : jump back to matching [ if byte at pointer is nonzero
//
// All other characters are treated as comments and ignored.
//
// IR instructions:
//
//	SHIFT k    ; move data pointer
//	ADD k      ; add to current cell (wraps mod 256)
//	ZERO       ; set current cell to 0
//	IN         ; read byte into cell
//	OUT        ; write byte from cell
//	JZ target  ; conditional jump if cell == 0
//	JNZ target ; conditional jump if cell != 0
package core

// TapeSize is the size of the Brainfuck tape in bytes (traditional 30KB).
const TapeSize = 30000

// Position represents a location in the source file.
type Position struct {
	Offset int // byte offset from start of file
	Line   int // 1-based line number
	Column int // 1-based column number
}
