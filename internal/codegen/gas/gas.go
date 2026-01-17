// Package gas provides GAS (GNU Assembler) assembly output for x86_64 Linux.
package gas

import (
	"fmt"
	"strings"

	"github.com/lcox74/bfcc/internal/core"
)

// Linux syscall numbers
const (
	sysWrite = 1
	sysExit  = 60
)

// TapeSize is the size of the Brainfuck tape in bytes.
const TapeSize = 30000

// Generator produces GAS (AT&T syntax) assembly from IR operations.
type Generator struct {
	ops     []core.Op
	out     strings.Builder
	targets map[int]bool
}

// NewGenerator creates a new GAS assembly generator.
func NewGenerator(ops []core.Op) *Generator {
	g := &Generator{ops: ops, targets: make(map[int]bool)}
	g.collectTargets()
	return g
}

// collectTargets finds all jump target indices.
func (g *Generator) collectTargets() {
	for _, op := range g.ops {
		if op.Kind == core.OpJz || op.Kind == core.OpJnz {
			g.targets[op.Arg] = true
		}
	}
}

// Generate produces the complete assembly output.
func (g *Generator) Generate() string {
	g.emitHeader()
	g.emitPrologue()

	for i, op := range g.ops {
		if g.targets[i] {
			g.emitLabel(i)
		}
		g.emitOp(op)
	}

	if g.targets[len(g.ops)] {
		g.emitLabel(len(g.ops))
	}
	g.emitEpilogue()
	g.emitHelpers()

	return g.out.String()
}

// emitHeader outputs the assembly file header with BSS and text sections.
func (g *Generator) emitHeader() {
	fmt.Fprintf(&g.out, ".section .bss\n")
	fmt.Fprintf(&g.out, "    .lcomm tape, %d\n", TapeSize)
	fmt.Fprintf(&g.out, "\n")
	fmt.Fprintf(&g.out, ".section .text\n")
	fmt.Fprintf(&g.out, ".globl _start\n")
}

// emitPrologue outputs the program start: initialize R13 (tape base) and R12 (data pointer).
func (g *Generator) emitPrologue() {
	fmt.Fprintf(&g.out, "_start:\n")

	// Load tape base address into R13
	fmt.Fprintf(&g.out, "    movq $tape, %%r13\n")

	// Zero the data pointer (R12)
	fmt.Fprintf(&g.out, "    xorq %%r12, %%r12\n")
}

// emitEpilogue outputs the exit(0) syscall.
func (g *Generator) emitEpilogue() {
	fmt.Fprintf(&g.out, "    movq $%d, %%rax\n", sysExit)
	fmt.Fprintf(&g.out, "    xorq %%rdi, %%rdi\n")
	fmt.Fprintf(&g.out, "    syscall\n")
}

// emitHelpers outputs the I/O helper functions.
func (g *Generator) emitHelpers() {
	fmt.Fprintf(&g.out, "\n_bf_read:\n")
	fmt.Fprintf(&g.out, "    leaq (%%r13,%%r12), %%rsi\n")
	fmt.Fprintf(&g.out, "    xorq %%rax, %%rax\n")
	fmt.Fprintf(&g.out, "    xorq %%rdi, %%rdi\n")
	fmt.Fprintf(&g.out, "    movq $1, %%rdx\n")
	fmt.Fprintf(&g.out, "    syscall\n")
	fmt.Fprintf(&g.out, "    ret\n")

	fmt.Fprintf(&g.out, "\n_bf_write:\n")
	fmt.Fprintf(&g.out, "    leaq (%%r13,%%r12), %%rsi\n")
	fmt.Fprintf(&g.out, "    movq $%d, %%rax\n", sysWrite)
	fmt.Fprintf(&g.out, "    movq $1, %%rdi\n")
	fmt.Fprintf(&g.out, "    movq $1, %%rdx\n")
	fmt.Fprintf(&g.out, "    syscall\n")
	fmt.Fprintf(&g.out, "    ret\n")
}

// emitLabel outputs a label for the given IR index.
func (g *Generator) emitLabel(index int) {
	fmt.Fprintf(&g.out, ".jt_%d:\n", index)
}

// emitOp outputs assembly for a single IR operation.
func (g *Generator) emitOp(op core.Op) {
	switch op.Kind {
	case core.OpShift:
		g.emitShift(op.Arg)
	case core.OpAdd:
		g.emitAdd(op.Arg)
	case core.OpZero:
		g.emitZero()
	case core.OpIn:
		g.emitIn()
	case core.OpOut:
		g.emitOut()
	case core.OpJz:
		g.emitJz(op.Arg)
	case core.OpJnz:
		g.emitJnz(op.Arg)
	}
}

// emitShift outputs: addq $k, %r12 (or subq for negative values)
func (g *Generator) emitShift(k int) {
	if k == 0 {
		return
	}
	if k > 0 {
		fmt.Fprintf(&g.out, "    addq $%d, %%r12\n", k)
	} else {
		fmt.Fprintf(&g.out, "    subq $%d, %%r12\n", -k)
	}
}

// emitAdd outputs: addb $k, (%r13,%r12) (or subb for negative values)
func (g *Generator) emitAdd(k int) {
	if k == 0 {
		return
	}
	if k > 0 {
		fmt.Fprintf(&g.out, "    addb $%d, (%%r13,%%r12)\n", k)
	} else {
		fmt.Fprintf(&g.out, "    subb $%d, (%%r13,%%r12)\n", -k)
	}
}

// emitZero outputs: movb $0, (%r13,%r12)
func (g *Generator) emitZero() {
	fmt.Fprintf(&g.out, "    movb $0, (%%r13,%%r12)\n")
}

// emitIn outputs a call to the read helper.
func (g *Generator) emitIn() {
	fmt.Fprintf(&g.out, "    call _bf_read\n")
}

// emitOut outputs a call to the write helper.
func (g *Generator) emitOut() {
	fmt.Fprintf(&g.out, "    call _bf_write\n")
}

// emitJz outputs: testb $0xff, (%r13,%r12); jz target
func (g *Generator) emitJz(target int) {
	fmt.Fprintf(&g.out, "    testb $0xff, (%%r13,%%r12)\n")
	fmt.Fprintf(&g.out, "    jz .jt_%d\n", target)
}

// emitJnz outputs: testb $0xff, (%r13,%r12); jnz target
func (g *Generator) emitJnz(target int) {
	fmt.Fprintf(&g.out, "    testb $0xff, (%%r13,%%r12)\n")
	fmt.Fprintf(&g.out, "    jnz .jt_%d\n", target)
}
