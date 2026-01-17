package amd64

// This file contains x86_64 instruction encoders.
// Each function returns the machine code bytes for a specific instruction.
//
// For details on x86-64 instruction encoding (REX prefixes, ModRM, SIB bytes),
// see: https://wiki.osdev.org/X86-64_Instruction_Encoding
//
// It was late and the level of headaches were growing, so I had this file
// generated based on that information and the gas instructions that I needed.

// MovabsR13 encodes: movabs $imm64, %r13 (49 BD <imm64>)
// Loads a 64-bit immediate into R13.
func MovabsR13(imm64 uint64) []byte {
	// REX.WB (49) = REX.W (64-bit) + REX.B (R13)
	// B8+r = mov imm64 to register, with R13: BD
	buf := make([]byte, 10)
	buf[0] = 0x49 // REX.WB
	buf[1] = 0xBD // mov r13, imm64
	writeLE64(buf[2:], imm64)
	return buf
}

// XorR12R12 encodes: xorq %r12, %r12 (4D 31 E4)
// Zeros R12.
func XorR12R12() []byte {
	// REX.WRB (4D) = REX.W + REX.R (r12 in reg) + REX.B (r12 in rm)
	// 31 /r = xor r/m64, r64
	// ModRM: 11 (reg-reg) 100 (r12) 100 (r12) = E4
	return []byte{0x4D, 0x31, 0xE4}
}

// AddqImm32R12 encodes: addq $imm32, %r12 (49 81 C4 <imm32>)
// Adds a signed 32-bit immediate to R12.
func AddqImm32R12(imm32 int32) []byte {
	// REX.WB (49) = REX.W + REX.B (R12)
	// 81 /0 id = add r/m64, imm32
	// ModRM: 11 (reg) 000 (/0) 100 (r12) = C4
	buf := make([]byte, 7)
	buf[0] = 0x49
	buf[1] = 0x81
	buf[2] = 0xC4
	writeLE32(buf[3:], uint32(imm32))
	return buf
}

// SubqImm32R12 encodes: subq $imm32, %r12 (49 81 EC <imm32>)
// Subtracts a signed 32-bit immediate from R12.
func SubqImm32R12(imm32 int32) []byte {
	// REX.WB (49) = REX.W + REX.B (R12)
	// 81 /5 id = sub r/m64, imm32
	// ModRM: 11 (reg) 101 (/5) 100 (r12) = EC
	buf := make([]byte, 7)
	buf[0] = 0x49
	buf[1] = 0x81
	buf[2] = 0xEC
	writeLE32(buf[3:], uint32(imm32))
	return buf
}

// AddbImm8Mem encodes: addb $imm8, (%r13,%r12) (43 80 44 25 00 <imm8>)
// Adds an unsigned 8-bit immediate to the byte at (%r13,%r12).
func AddbImm8Mem(imm8 uint8) []byte {
	// 43 = REX.XB (X for r12 in SIB.index, B for r13 in SIB.base)
	// 80 /0 ib = add r/m8, imm8
	// ModRM: 01 (disp8) 000 (/0) 100 (SIB) = 44
	// SIB: 00 (scale=1) 100 (r12 index) 101 (r13 base) = 25
	// disp8 = 00 (required due to r13 base encoding)
	return []byte{0x43, 0x80, 0x44, 0x25, 0x00, imm8}
}

// SubbImm8Mem encodes: subb $imm8, (%r13,%r12) (43 80 6C 25 00 <imm8>)
// Subtracts an unsigned 8-bit immediate from the byte at (%r13,%r12).
func SubbImm8Mem(imm8 uint8) []byte {
	// 43 = REX.XB
	// 80 /5 ib = sub r/m8, imm8
	// ModRM: 01 (disp8) 101 (/5) 100 (SIB) = 6C
	// SIB: 00 (scale=1) 100 (r12 index) 101 (r13 base) = 25
	// disp8 = 00
	return []byte{0x43, 0x80, 0x6C, 0x25, 0x00, imm8}
}

// MovbZeroMem encodes: movb $0, (%r13,%r12) (43 C6 44 25 00 00)
// Sets the byte at (%r13,%r12) to 0.
func MovbZeroMem() []byte {
	// 43 = REX.XB
	// C6 /0 ib = mov r/m8, imm8
	// ModRM: 01 (disp8) 000 (/0) 100 (SIB) = 44
	// SIB: 00 (scale=1) 100 (r12 index) 101 (r13 base) = 25
	// disp8 = 00, imm8 = 00
	return []byte{0x43, 0xC6, 0x44, 0x25, 0x00, 0x00}
}

// TestbMem encodes: testb $0xff, (%r13,%r12) (43 F6 44 25 00 FF)
// Tests the byte at (%r13,%r12) against 0xFF, setting flags.
func TestbMem() []byte {
	// 43 = REX.XB
	// F6 /0 ib = test r/m8, imm8
	// ModRM: 01 (disp8) 000 (/0) 100 (SIB) = 44
	// SIB: 00 (scale=1) 100 (r12 index) 101 (r13 base) = 25
	// disp8 = 00, imm8 = FF
	return []byte{0x43, 0xF6, 0x44, 0x25, 0x00, 0xFF}
}

// JzRel32 encodes: jz rel32 (0F 84 <rel32>)
// Jump if zero flag is set. rel32 is relative to end of instruction.
func JzRel32(rel32 int32) []byte {
	buf := make([]byte, 6)
	buf[0] = 0x0F
	buf[1] = 0x84
	writeLE32(buf[2:], uint32(rel32))
	return buf
}

// JnzRel32 encodes: jnz rel32 (0F 85 <rel32>)
// Jump if zero flag is not set. rel32 is relative to end of instruction.
func JnzRel32(rel32 int32) []byte {
	buf := make([]byte, 6)
	buf[0] = 0x0F
	buf[1] = 0x85
	writeLE32(buf[2:], uint32(rel32))
	return buf
}

// CallRel32 encodes: call rel32 (E8 <rel32>)
// Call a function. rel32 is relative to end of instruction.
func CallRel32(rel32 int32) []byte {
	buf := make([]byte, 5)
	buf[0] = 0xE8
	writeLE32(buf[1:], uint32(rel32))
	return buf
}

// Ret encodes: ret (C3)
func Ret() []byte {
	return []byte{0xC3}
}

// Syscall encodes: syscall (0F 05)
func Syscall() []byte {
	return []byte{0x0F, 0x05}
}

// LeaqR13R12ToRSI encodes: leaq (%r13,%r12), %rsi (4B 8D 74 25 00)
// Load effective address of (%r13,%r12) into RSI.
func LeaqR13R12ToRSI() []byte {
	// 4B = REX.WXB (W=64-bit, X=r12 in SIB.index, B=r13 in SIB.base)
	// 8D /r = lea r64, m
	// ModRM: 01 (disp8) 110 (rsi) 100 (SIB) = 74
	// SIB: 00 (scale=1) 100 (r12 index) 101 (r13 base) = 25
	// disp8 = 00
	return []byte{0x4B, 0x8D, 0x74, 0x25, 0x00}
}

// XorRAXRAX encodes: xorq %rax, %rax (48 31 C0)
// Zeros RAX.
func XorRAXRAX() []byte {
	return []byte{0x48, 0x31, 0xC0}
}

// XorRDIRDI encodes: xorq %rdi, %rdi (48 31 FF)
// Zeros RDI.
func XorRDIRDI() []byte {
	return []byte{0x48, 0x31, 0xFF}
}

// MovqImm32RAX encodes: movq $imm32, %rax (48 C7 C0 <imm32>)
// Load 32-bit sign-extended immediate into RAX.
func MovqImm32RAX(imm32 int32) []byte {
	buf := make([]byte, 7)
	buf[0] = 0x48 // REX.W
	buf[1] = 0xC7 // mov r/m64, imm32
	buf[2] = 0xC0 // ModRM: 11 000 000 (rax)
	writeLE32(buf[3:], uint32(imm32))
	return buf
}

// MovqImm32RDI encodes: movq $imm32, %rdi (48 C7 C7 <imm32>)
// Load 32-bit sign-extended immediate into RDI.
func MovqImm32RDI(imm32 int32) []byte {
	buf := make([]byte, 7)
	buf[0] = 0x48 // REX.W
	buf[1] = 0xC7 // mov r/m64, imm32
	buf[2] = 0xC7 // ModRM: 11 000 111 (rdi)
	writeLE32(buf[3:], uint32(imm32))
	return buf
}

// MovqImm32RDX encodes: movq $imm32, %rdx (48 C7 C2 <imm32>)
// Load 32-bit sign-extended immediate into RDX.
func MovqImm32RDX(imm32 int32) []byte {
	buf := make([]byte, 7)
	buf[0] = 0x48 // REX.W
	buf[1] = 0xC7 // mov r/m64, imm32
	buf[2] = 0xC2 // ModRM: 11 000 010 (rdx)
	writeLE32(buf[3:], uint32(imm32))
	return buf
}
