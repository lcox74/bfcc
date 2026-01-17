package vm

import (
	"fmt"

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
