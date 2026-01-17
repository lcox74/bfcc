// Package vm provides a Brainfuck interpreter for executing IR operations.
package vm

import (
	"fmt"
	"io"
	"os"

	"github.com/lcox74/bfcc/internal/core"
)

// RuntimeError represents an error during VM execution.
type RuntimeError struct {
	Msg string
	Pos *core.Position
	PC  int
}

func (e *RuntimeError) Error() string {
	if e.Pos != nil {
		return fmt.Sprintf("runtime error at PC %d (line %d, col %d): %s",
			e.PC,
			e.Pos.Line,
			e.Pos.Column,
			e.Msg,
		)
	}
	return fmt.Sprintf("runtime error at PC %d: %s", e.PC, e.Msg)
}

// EOFBehavior specifies how the VM handles EOF on input.
type EOFBehavior int

const (
	EOFZero     EOFBehavior = iota // Set cell to 0 (default)
	EOFMinusOne                    // Set cell to 255
	EOFNoChange                    // Leave cell unchanged
)

// VM executes Brainfuck IR operations.
type VM struct {
	memSize     int
	input       io.Reader
	output      io.Writer
	eofBehavior EOFBehavior
	memory      []byte
	dp          int     // data pointer
	pc          int     // program counter
	ioBuf       [1]byte // reusable I/O buffer to avoid allocations
}

// VMOption is a functional option for configuring a VM.
type VMOption func(*VM)

// WithMemorySize sets the memory size (default 30000).
func WithMemorySize(size int) VMOption {
	return func(v *VM) {
		v.memSize = size
	}
}

// WithInput sets the input reader (default os.Stdin).
func WithInput(r io.Reader) VMOption {
	return func(v *VM) {
		v.input = r
	}
}

// WithOutput sets the output writer (default os.Stdout).
func WithOutput(w io.Writer) VMOption {
	return func(v *VM) {
		v.output = w
	}
}

// WithEOFBehavior sets the EOF handling behavior (default EOFZero).
func WithEOFBehavior(b EOFBehavior) VMOption {
	return func(v *VM) {
		v.eofBehavior = b
	}
}

// NewVM creates a new VM with the given options.
func NewVM(opts ...VMOption) *VM {
	vm := &VM{
		memSize:     30000,
		input:       os.Stdin,
		output:      os.Stdout,
		eofBehavior: EOFZero,
	}
	for _, opt := range opts {
		opt(vm)
	}
	return vm
}

// Run executes the given IR operations.
func (v *VM) Run(ops []core.Op) error {
	v.memory = make([]byte, v.memSize)
	v.dp = 0
	v.pc = 0

	// Cache frequently accessed values for the hot loop
	memory := v.memory
	memSize := v.memSize
	numOps := len(ops)

	for v.pc < numOps {
		op := ops[v.pc]

		switch op.Kind {
		case core.OpShift:
			v.dp += op.Arg
			if v.dp < 0 || v.dp >= memSize {
				return &RuntimeError{
					Msg: fmt.Sprintf("data pointer out of bounds: %d (valid range 0-%d)", v.dp, memSize-1),
					Pos: op.Pos,
					PC:  v.pc,
				}
			}

		case core.OpAdd:
			memory[v.dp] += byte(op.Arg)

		case core.OpZero:
			memory[v.dp] = 0

		case core.OpIn:
			n, err := v.input.Read(v.ioBuf[:])
			if err == io.EOF || n == 0 {
				switch v.eofBehavior {
				case EOFZero:
					memory[v.dp] = 0
				case EOFMinusOne:
					memory[v.dp] = 255
				case EOFNoChange:
					// leave unchanged
				}
			} else if err != nil {
				return &RuntimeError{
					Msg: fmt.Sprintf("input error: %v", err),
					Pos: op.Pos,
					PC:  v.pc,
				}
			} else {
				memory[v.dp] = v.ioBuf[0]
			}

		case core.OpOut:
			v.ioBuf[0] = memory[v.dp]
			_, err := v.output.Write(v.ioBuf[:])
			if err != nil {
				return &RuntimeError{
					Msg: fmt.Sprintf("output error: %v", err),
					Pos: op.Pos,
					PC:  v.pc,
				}
			}

		case core.OpJz:
			if memory[v.dp] == 0 {
				v.pc = op.Arg
				continue
			}

		case core.OpJnz:
			if memory[v.dp] != 0 {
				v.pc = op.Arg
				continue
			}
		}

		v.pc++
	}

	return nil
}
