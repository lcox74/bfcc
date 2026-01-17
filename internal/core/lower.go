package core

import "fmt"

// Error is returned when lowering fails (eg. unmatched brackets).
type Error struct {
	Msg string
	Pos Position
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s at line %d col %d (offset %d)",
		e.Msg, e.Pos.Line, e.Pos.Column, e.Pos.Offset)
}

// lowerRule describes how to lower a token kind to an IR op.
type lowerRule struct {
	op   OpKind
	sign int  // multiplier for foldable ops (+1 or -1)
	fold bool // true if consecutive tokens should be folded
}

// tokToRule maps foldable/simple token kinds to their lowering rules.
var tokToRule = [...]lowerRule{
	TokShiftRight: {OpShift, +1, true},
	TokShiftLeft:  {OpShift, -1, true},
	TokAdd:        {OpAdd, +1, true},
	TokSub:        {OpAdd, -1, true},
	TokOut:        {OpOut, 0, false},
	TokIn:         {OpIn, 0, false},
}

// Lower converts a token stream into IR operations.
func Lower(toks []Token) ([]Op, error) {
	ops := make([]Op, 0, len(toks))
	loopStack := make([]int, 0, 8)

	for i := 0; i < len(toks); {
		tok := toks[i]
		pos := &Position{tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column}

		switch tok.Kind {
		case TokEOF:
			if len(loopStack) > 0 {
				return nil, &Error{"unmatched '['", toks[loopStack[0]].Pos}
			}

			return ops, nil

		case TokLBracket:
			loopStack = append(loopStack, len(ops))
			ops = append(ops, Op{Kind: OpJz, Pos: pos})
			i++

		case TokRBracket:
			if len(loopStack) == 0 {
				return nil, &Error{"unmatched ']'", tok.Pos}
			}

			start := loopStack[len(loopStack)-1]
			loopStack = loopStack[:len(loopStack)-1]
			ops = append(ops, Op{Kind: OpJnz, Arg: start, Pos: pos})
			ops[start].Arg = len(ops)
			i++

		case TokAdd, TokSub, TokShiftLeft, TokShiftRight, TokIn, TokOut:
			rule := tokToRule[tok.Kind]
			if rule.fold {
				count := FoldToken(toks, i, tok.Kind)
				ops = append(ops, Op{Kind: rule.op, Arg: rule.sign * count, Pos: pos})
				i += count
				continue
			}

			ops = append(ops, Op{Kind: rule.op, Pos: pos})
			i++

		default:
			return nil, &Error{"unexpected token", tok.Pos}
		}
	}
	return ops, nil
}
