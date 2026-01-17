// Package linux produces ELF64 x86_64 Linux executables from IR operations.
package linux

import (
	"encoding/binary"

	"github.com/lcox74/bfcc/internal/core"
	"github.com/lcox74/bfcc/pkg/amd64"
	"github.com/lcox74/bfcc/pkg/elf"
)

// Linux syscall numbers
const (
	// sysRead = 0 // Omitted, it's quicker to use xor to zero out
	sysWrite = 1
	sysExit  = 60
)

// Memory layout constants
const (
	CodeBase = 0x400000 // Virtual address for code segment
	BSSBase  = 0x600000 // Virtual address for BSS segment (tape)
)

// jumpFixup records a location that needs to be patched with a relative offset.
type jumpFixup struct {
	offset    int // Offset in code where rel32 starts
	targetIdx int // IR index of the jump target
}

// X86_64Generator produces x86_64 machine code from IR operations.
type X86_64Generator struct {
	ops       []core.Op
	code      []byte
	targets   map[int]bool // IR indices that are jump targets
	labelAddr map[int]int  // IR index -> code offset
	fixups    []jumpFixup  // Jumps that need patching
	codeBase  uint64       // Virtual address where code will be loaded
	bssBase   uint64       // Virtual address for BSS/tape
}

// NewX86_64Generator creates a new x86_64 machine code generator.
func NewX86_64Generator(ops []core.Op) *X86_64Generator {
	g := &X86_64Generator{
		ops:       ops,
		code:      make([]byte, 0, 4096),
		targets:   make(map[int]bool),
		labelAddr: make(map[int]int),
		codeBase:  CodeBase + elf.PageSize, // Code starts after ELF headers
		bssBase:   BSSBase,
	}
	g.collectTargets()
	return g
}

// collectTargets finds all jump target indices.
func (g *X86_64Generator) collectTargets() {
	for _, op := range g.ops {
		if op.Kind == core.OpJz || op.Kind == core.OpJnz {
			g.targets[op.Arg] = true
		}
	}
}

// Generate produces raw x86_64 machine code.
func (g *X86_64Generator) Generate() []byte {
	g.emitPrologue()

	for i, op := range g.ops {
		if g.targets[i] {
			g.labelAddr[i] = len(g.code)
		}
		g.emitOp(op)
	}

	// Record final label address if it's a target
	if g.targets[len(g.ops)] {
		g.labelAddr[len(g.ops)] = len(g.code)
	}

	g.emitEpilogue()
	g.emitHelpers()
	g.resolveFixups()

	return g.code
}

// GenerateELF produces a complete ELF64 executable.
func (g *X86_64Generator) GenerateELF() []byte {
	code := g.Generate()

	builder := elf.NewBuilder()
	builder.SetEntry(g.codeBase)
	builder.AddLoadSegment(code, g.codeBase, elf.PF_R|elf.PF_X)
	builder.AddBSSSegment(g.bssBase, core.TapeSize, elf.PF_R|elf.PF_W)

	return builder.Build()
}

// emitBytes appends a byte slice to the code buffer.
func (g *X86_64Generator) emitBytes(b []byte) {
	g.code = append(g.code, b...)
}

// emitPrologue outputs the program start: initialize R13 (tape base) and R12 (data pointer).
func (g *X86_64Generator) emitPrologue() {
	// Load tape base address
	g.emitBytes(amd64.MovabsR13(g.bssBase)) // movabs $tape, %r13

	// Zero data pointer
	g.emitBytes(amd64.XorR12R12()) // xorq %r12, %r12
}

// emitEpilogue outputs the exit(0) syscall.
func (g *X86_64Generator) emitEpilogue() {
	// Set Exit syscall
	g.emitBytes(amd64.MovqImm32RAX(sysExit)) // mov $60, %rax

	// Set Exit code 0
	g.emitBytes(amd64.XorRDIRDI()) // xor %rdi, %rdi

	// Perform Syscall
	g.emitBytes(amd64.Syscall()) // syscall
}

// helperReadOffset and helperWriteOffset store the code offsets of helper functions.
var helperReadOffset, helperWriteOffset int

// emitHelpers outputs the I/O helper functions.
func (g *X86_64Generator) emitHelpers() {
	// _bf_read:
	helperReadOffset = len(g.code)
	g.emitBytes(amd64.LeaqR13R12ToRSI()) // leaq (%r13,%r12), %rsi
	g.emitBytes(amd64.XorRAXRAX())       // xorq %rax, %rax - syscall 0 (read)
	g.emitBytes(amd64.XorRDIRDI())       // xorq %rdi, %rdi
	g.emitBytes(amd64.MovqImm32RDX(1))   // movq $1, %rdx
	g.emitBytes(amd64.Syscall())         // syscall
	g.emitBytes(amd64.Ret())             // ret

	// _bf_write:
	helperWriteOffset = len(g.code)
	g.emitBytes(amd64.LeaqR13R12ToRSI())      // leaq (%r13,%r12), %rsi
	g.emitBytes(amd64.MovqImm32RAX(sysWrite)) // movq $1, %rax - syscall 1 (write)
	g.emitBytes(amd64.MovqImm32RDI(1))        // movq $1, %rdi
	g.emitBytes(amd64.MovqImm32RDX(1))        // movq $1, %rdx
	g.emitBytes(amd64.Syscall())              // syscall
	g.emitBytes(amd64.Ret())                  // ret
}

// emitOp outputs machine code for a single IR operation.
func (g *X86_64Generator) emitOp(op core.Op) {
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

// emitShift outputs: addq/subq $k, %r12
// Uses 32-bit immediate since Op.Arg is int.
func (g *X86_64Generator) emitShift(k int) {
	if k == 0 {
		return
	}
	if k > 0 {
		g.emitBytes(amd64.AddqImm32R12(int32(k))) // addq $k, %r12
	} else {
		g.emitBytes(amd64.SubqImm32R12(int32(-k))) // subq $k, %r12
	}
}

// emitAdd outputs: addb/subb $k, (%r13,%r12)
// Tape cells are unsigned bytes [0, 255], so we use separate add/sub with uint8 immediates.
func (g *X86_64Generator) emitAdd(k int) {
	if k == 0 {
		return
	}
	if k > 0 {
		g.emitBytes(amd64.AddbImm8Mem(uint8(k))) // addb $k, (%r13,%r12)
	} else {
		g.emitBytes(amd64.SubbImm8Mem(uint8(-k))) // subb $k, (%r13,%r12)
	}
}

// emitZero outputs: movb $0, (%r13,%r12)
func (g *X86_64Generator) emitZero() {
	g.emitBytes(amd64.MovbZeroMem()) // movb $0, (%r13,%r12)
}

// emitIn outputs a call to _bf_read helper.
func (g *X86_64Generator) emitIn() {
	// Placeholder call - will be fixed up after helpers are emitted
	g.fixups = append(g.fixups, jumpFixup{
		offset:    len(g.code) + 1, // rel32 starts at offset 1 in call instruction
		targetIdx: -1,              // Special marker for read helper
	})
	g.emitBytes(amd64.CallRel32(0)) // Placeholder
}

// emitOut outputs a call to _bf_write helper.
func (g *X86_64Generator) emitOut() {
	// Placeholder call - will be fixed up after helpers are emitted
	g.fixups = append(g.fixups, jumpFixup{
		offset:    len(g.code) + 1, // rel32 starts at offset 1 in call instruction
		targetIdx: -2,              // Special marker for write helper
	})
	g.emitBytes(amd64.CallRel32(0)) // Placeholder
}

// emitJz outputs: testb $0xff, (%r13,%r12); jz target
func (g *X86_64Generator) emitJz(target int) {
	g.emitBytes(amd64.TestbMem())
	// Record fixup for the jz rel32
	g.fixups = append(g.fixups, jumpFixup{
		offset:    len(g.code) + 2, // rel32 starts at offset 2 in jz instruction
		targetIdx: target,
	})
	g.emitBytes(amd64.JzRel32(0)) // Placeholder
}

// emitJnz outputs: testb $0xff, (%r13,%r12); jnz target
func (g *X86_64Generator) emitJnz(target int) {
	g.emitBytes(amd64.TestbMem())
	// Record fixup for the jnz rel32
	g.fixups = append(g.fixups, jumpFixup{
		offset:    len(g.code) + 2, // rel32 starts at offset 2 in jnz instruction
		targetIdx: target,
	})
	g.emitBytes(amd64.JnzRel32(0)) // Placeholder
}

// resolveFixups patches all jump and call targets.
func (g *X86_64Generator) resolveFixups() {
	for _, fixup := range g.fixups {
		var targetAddr int
		switch fixup.targetIdx {
		case -1: // read helper
			targetAddr = helperReadOffset
		case -2: // write helper
			targetAddr = helperWriteOffset
		default:
			targetAddr = g.labelAddr[fixup.targetIdx]
		}

		// Calculate relative offset from end of instruction
		// For jz/jnz: instruction ends 4 bytes after rel32 start
		// For call: instruction ends 4 bytes after rel32 start
		instrEnd := fixup.offset + 4
		rel32 := int32(targetAddr - instrEnd)

		// Patch the rel32 in place
		binary.LittleEndian.PutUint32(g.code[fixup.offset:], uint32(rel32))
	}
}
