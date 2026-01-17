package core

// Optimise applies peephole and structural optimisations to the IR.
// It returns a new slice with the optimised operations.
func Optimise(ops []Op) []Op {
	if len(ops) == 0 {
		return ops
	}

	// Apply optimisation passes until no more changes occur
	result := ops
	for {
		prev := len(result)
		result = clearLoops(result)
		result = removeEmptyLoops(result)
		result = mergeAdjacent(result)
		result = removeNoOps(result)
		if len(result) == prev {
			break
		}
	}

	return result
}

// removeEmptyLoops eliminates empty [] loops (JZ immediately followed by JNZ).
// These are often used as comments in Brainfuck: [this is a comment]
func removeEmptyLoops(ops []Op) []Op {
	if len(ops) < 2 {
		return ops
	}

	result := make([]Op, 0, len(ops))
	i := 0

	for i < len(ops) {
		// Check for empty loop: JZ immediately followed by its matching JNZ
		if i+1 < len(ops) &&
			ops[i].Kind == OpJz &&
			ops[i+1].Kind == OpJnz &&
			ops[i].Arg == i+2 &&
			ops[i+1].Arg == i {
			// Skip both instructions
			i += 2
			continue
		}

		result = append(result, ops[i])
		i++
	}

	return fixJumpTargets(result)
}

// clearLoops detects [-] and [+] patterns and replaces them with ZERO.
// Pattern: JZ target, ADD ±1, JNZ start (where target = start+3, JNZ points to start)
func clearLoops(ops []Op) []Op {
	if len(ops) < 3 {
		return ops
	}

	result := make([]Op, 0, len(ops))
	i := 0

	for i < len(ops) {
		// Check for clear loop pattern: JZ, ADD ±1, JNZ
		if i+2 < len(ops) &&
			ops[i].Kind == OpJz &&
			ops[i+1].Kind == OpAdd &&
			(ops[i+1].Arg == 1 || ops[i+1].Arg == -1) &&
			ops[i+2].Kind == OpJnz &&
			ops[i].Arg == i+3 &&
			ops[i+2].Arg == i {
			// Replace with ZERO, preserving position from the opening bracket
			result = append(result, Op{Kind: OpZero, Pos: ops[i].Pos})
			i += 3
			continue
		}

		result = append(result, ops[i])
		i++
	}

	// Fix up jump targets after removing instructions
	return fixJumpTargets(result)
}

// mergeAdjacent combines consecutive ADD or SHIFT operations.
func mergeAdjacent(ops []Op) []Op {
	if len(ops) < 2 {
		return ops
	}

	result := make([]Op, 0, len(ops))

	for _, op := range ops {
		if len(result) == 0 {
			result = append(result, op)
			continue
		}

		last := &result[len(result)-1]

		// Merge consecutive ADD operations
		if op.Kind == OpAdd && last.Kind == OpAdd {
			last.Arg += op.Arg
			continue
		}

		// Merge consecutive SHIFT operations
		if op.Kind == OpShift && last.Kind == OpShift {
			last.Arg += op.Arg
			continue
		}

		result = append(result, op)
	}

	// Fix up jump targets after merging instructions
	return fixJumpTargets(result)
}

// removeNoOps eliminates operations that have no effect and normalizes ADD values.
func removeNoOps(ops []Op) []Op {
	result := make([]Op, 0, len(ops))

	for _, op := range ops {
		// Normalize ADD to [-255, 255] range (8-bit cells)
		if op.Kind == OpAdd {
			op.Arg = op.Arg % 256
		}

		// Skip ADD 0 and SHIFT 0
		if (op.Kind == OpAdd || op.Kind == OpShift) && op.Arg == 0 {
			continue
		}

		result = append(result, op)
	}

	// Fix up jump targets after removing instructions
	return fixJumpTargets(result)
}

// fixJumpTargets recalculates JZ/JNZ targets after instructions are removed.
// Uses bracket matching to pair JZ with corresponding JNZ.
func fixJumpTargets(ops []Op) []Op {
	// Build a mapping of loop pairs
	stack := make([]int, 0, 8)

	for i, op := range ops {
		switch op.Kind {
		case OpJz:
			stack = append(stack, i)
		case OpJnz:
			if len(stack) > 0 {
				start := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				// JZ jumps past the JNZ
				ops[start].Arg = i + 1
				// JNZ jumps back to the JZ
				ops[i].Arg = start
			}
		}
	}

	return ops
}
