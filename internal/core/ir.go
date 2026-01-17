package core

import (
	"fmt"
	"strings"
)

// OpKind identifies the kind of IR operation.
type OpKind int

const (
	OpShift OpKind = iota // SHIFT k
	OpAdd                 // ADD k
	OpZero                // ZERO
	OpIn                  // IN
	OpOut                 // OUT
	OpJz                  // JZ target
	OpJnz                 // JNZ target
)

// opNames maps each OpKind to its string representation for debugging.
var opNames = [...]string{
	OpShift: "SHIFT",
	OpAdd:   "ADD",
	OpZero:  "ZERO",
	OpIn:    "IN",
	OpOut:   "OUT",
	OpJz:    "JZ",
	OpJnz:   "JNZ",
}

// String returns the string representation of the OpKind.
func (k OpKind) String() string {
	return opNames[k]
}

// Op represents one intermediate instruction.
type Op struct {
	Kind OpKind
	Arg  int       // used by SHIFT/ADD/JZ/JNZ
	Pos  *Position // optional source metadata for debugging
}

func Shift(k int) Op    { return Op{Kind: OpShift, Arg: k} }
func Add(k int) Op      { return Op{Kind: OpAdd, Arg: k} }
func Zero() Op          { return Op{Kind: OpZero} }
func In() Op            { return Op{Kind: OpIn} }
func Out() Op           { return Op{Kind: OpOut} }
func Jz(target int) Op  { return Op{Kind: OpJz, Arg: target} }
func Jnz(target int) Op { return Op{Kind: OpJnz, Arg: target} }

// Dump returns a formatted string representation of the IR stream.
func Dump(ops []Op) string {
	var out strings.Builder

	for i, op := range ops {
		switch op.Kind {
		case OpShift:
			fmt.Fprintf(&out, "%03d: SHIFT %+d\n", i, op.Arg)
		case OpAdd:
			fmt.Fprintf(&out, "%03d: ADD   %+d\n", i, op.Arg)
		case OpZero:
			fmt.Fprintf(&out, "%03d: ZERO\n", i)
		case OpIn:
			fmt.Fprintf(&out, "%03d: IN\n", i)
		case OpOut:
			fmt.Fprintf(&out, "%03d: OUT\n", i)
		case OpJz:
			fmt.Fprintf(&out, "%03d: JZ    %d\n", i, op.Arg)
		case OpJnz:
			fmt.Fprintf(&out, "%03d: JNZ   %d\n", i, op.Arg)
		}
	}
	return out.String()
}
