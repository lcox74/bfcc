// Package elf provides ELF64 binary format building utilities.
// This package has no dependencies on the compiler internals and can be used
// standalone for generating ELF executables.
package elf

import (
	"encoding/binary"
)

// ELF64 constants
const (
	// ELF identification
	ELFMAG0       = 0x7f
	ELFMAG1       = 'E'
	ELFMAG2       = 'L'
	ELFMAG3       = 'F'
	ELFCLASS64    = 2
	ELFDATA2LSB   = 1 // Little endian
	EV_CURRENT    = 1
	ELFOSABI_NONE = 0

	// ELF types
	ET_EXEC = 2 // Executable file

	// Machine types
	EM_X86_64 = 62

	// Program header types
	PT_NULL = 0
	PT_LOAD = 1

	// Program header flags
	PF_X = 0x1 // Execute
	PF_W = 0x2 // Write
	PF_R = 0x4 // Read

	// Sizes
	ELF64HeaderSize = 64
	ELF64PhdrSize   = 56
	PageSize        = 0x1000
	DefaultCodeBase = 0x400000
	DefaultBSSBase  = 0x600000
)

// Header64 represents the ELF64 file header.
type Header64 struct {
	Ident     [16]byte // ELF identification
	Type      uint16   // Object file type
	Machine   uint16   // Machine type
	Version   uint32   // Object file version
	Entry     uint64   // Entry point address
	PhOff     uint64   // Program header offset
	ShOff     uint64   // Section header offset
	Flags     uint32   // Processor-specific flags
	EhSize    uint16   // ELF header size
	PhEntSize uint16   // Program header entry size
	PhNum     uint16   // Number of program headers
	ShEntSize uint16   // Section header entry size
	ShNum     uint16   // Number of section headers
	ShStrNdx  uint16   // Section name string table index
}

// Phdr64 represents an ELF64 program header.
type Phdr64 struct {
	Type   uint32 // Segment type
	Flags  uint32 // Segment flags
	Off    uint64 // File offset
	VAddr  uint64 // Virtual address
	PAddr  uint64 // Physical address
	FileSz uint64 // Size in file
	MemSz  uint64 // Size in memory
	Align  uint64 // Alignment
}

// Segment represents a loadable segment to be added to the ELF.
type Segment struct {
	VAddr uint64 // Virtual address
	Data  []byte // Segment data (nil for BSS)
	MemSz uint64 // Memory size (can be larger than len(Data) for BSS)
	Flags uint32 // PF_R, PF_W, PF_X
	IsBSS bool   // True if this is a BSS segment (no file data)
}

// Builder constructs an ELF64 executable.
type Builder struct {
	entry    uint64
	segments []Segment
}

// NewBuilder creates a new ELF64 builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// SetEntry sets the entry point virtual address.
func (b *Builder) SetEntry(vaddr uint64) {
	b.entry = vaddr
}

// AddLoadSegment adds a loadable segment with data.
func (b *Builder) AddLoadSegment(data []byte, vaddr uint64, flags uint32) {
	b.segments = append(b.segments, Segment{
		VAddr: vaddr,
		Data:  data,
		MemSz: uint64(len(data)),
		Flags: flags,
	})
}

// AddBSSSegment adds a BSS segment (zero-initialized, no file data).
func (b *Builder) AddBSSSegment(vaddr uint64, size uint64, flags uint32) {
	b.segments = append(b.segments, Segment{
		VAddr: vaddr,
		MemSz: size,
		Flags: flags,
		IsBSS: true,
	})
}

// Build produces the final ELF binary.
func (b *Builder) Build() []byte {
	// Calculate sizes
	numPhdrs := len(b.segments)
	headerSize := ELF64HeaderSize + numPhdrs*ELF64PhdrSize

	// Align code start to page boundary
	codeOffset := alignUp(uint64(headerSize), PageSize)

	// Build the binary
	out := make([]byte, 0, codeOffset)

	// Write ELF header
	out = b.writeHeader(out, numPhdrs)

	// Write program headers
	fileOffset := codeOffset
	for _, seg := range b.segments {
		var phdr Phdr64
		phdr.Type = PT_LOAD
		phdr.Flags = seg.Flags
		phdr.VAddr = seg.VAddr
		phdr.PAddr = seg.VAddr
		phdr.Align = PageSize

		if seg.IsBSS {
			// BSS: no file data, kernel zero-initializes
			phdr.Off = 0
			phdr.FileSz = 0
			phdr.MemSz = seg.MemSz
		} else {
			phdr.Off = fileOffset
			phdr.FileSz = uint64(len(seg.Data))
			phdr.MemSz = seg.MemSz
			fileOffset += uint64(len(seg.Data))
		}

		out = writePhdr(out, &phdr)
	}

	// Pad to code offset
	for len(out) < int(codeOffset) {
		out = append(out, 0)
	}

	// Write segment data
	for _, seg := range b.segments {
		if !seg.IsBSS {
			out = append(out, seg.Data...)
		}
	}

	return out
}

// writeHeader writes the ELF64 header.
//
//	ELF Layout (Minimal)
//
//	Offset     Content                Size
//	0x0000     ELF Header             64 bytes
//	0x0040     Program Header 1       56 bytes (PT_LOAD: code, R+X)
//	0x0078     Program Header 2       56 bytes (PT_LOAD: BSS, R+W)
//	0x1000     Code segment           variable (page-aligned)
//
//	Virtual Addresses:
//	0x400000   Code (mapped from file)
//	0x600000   BSS/tape (30KB, zero-initialized by kernel)
//
//	No section headers needed - just program headers for a minimal executable.
func (b *Builder) writeHeader(out []byte, numPhdrs int) []byte {
	hdr := Header64{
		Type:      ET_EXEC,
		Machine:   EM_X86_64,
		Version:   EV_CURRENT,
		Entry:     b.entry,
		PhOff:     ELF64HeaderSize,
		ShOff:     0, // No section headers
		Flags:     0,
		EhSize:    ELF64HeaderSize,
		PhEntSize: ELF64PhdrSize,
		PhNum:     uint16(numPhdrs),
		ShEntSize: 0,
		ShNum:     0,
		ShStrNdx:  0,
	}

	// ELF identification
	hdr.Ident[0] = ELFMAG0
	hdr.Ident[1] = ELFMAG1
	hdr.Ident[2] = ELFMAG2
	hdr.Ident[3] = ELFMAG3
	hdr.Ident[4] = ELFCLASS64
	hdr.Ident[5] = ELFDATA2LSB
	hdr.Ident[6] = EV_CURRENT
	hdr.Ident[7] = ELFOSABI_NONE
	// Ident[8..15] are padding (already zero)

	// Write header bytes
	out = append(out, hdr.Ident[:]...)
	out = appendLE16(out, hdr.Type)
	out = appendLE16(out, hdr.Machine)
	out = appendLE32(out, hdr.Version)
	out = appendLE64(out, hdr.Entry)
	out = appendLE64(out, hdr.PhOff)
	out = appendLE64(out, hdr.ShOff)
	out = appendLE32(out, hdr.Flags)
	out = appendLE16(out, hdr.EhSize)
	out = appendLE16(out, hdr.PhEntSize)
	out = appendLE16(out, hdr.PhNum)
	out = appendLE16(out, hdr.ShEntSize)
	out = appendLE16(out, hdr.ShNum)
	out = appendLE16(out, hdr.ShStrNdx)

	return out
}

// writePhdr writes a program header.
func writePhdr(out []byte, phdr *Phdr64) []byte {
	out = appendLE32(out, phdr.Type)
	out = appendLE32(out, phdr.Flags)
	out = appendLE64(out, phdr.Off)
	out = appendLE64(out, phdr.VAddr)
	out = appendLE64(out, phdr.PAddr)
	out = appendLE64(out, phdr.FileSz)
	out = appendLE64(out, phdr.MemSz)
	out = appendLE64(out, phdr.Align)
	return out
}

// Little-endian append helpers
func appendLE16(out []byte, v uint16) []byte {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	return append(out, buf[:]...)
}

func appendLE32(out []byte, v uint32) []byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	return append(out, buf[:]...)
}

func appendLE64(out []byte, v uint64) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	return append(out, buf[:]...)
}

func alignUp(v, align uint64) uint64 {
	return (v + align - 1) &^ (align - 1)
}
