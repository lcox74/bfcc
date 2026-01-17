// Package amd64 provides x86_64 (AMD64) machine code encoding utilities.
// This package has no dependencies on compiler internals and can be used
// standalone for generating x86_64 machine code.
package amd64

import "encoding/binary"

// writeLE32 writes a 32-bit value in little-endian order.
func writeLE32(buf []byte, v uint32) {
	binary.LittleEndian.PutUint32(buf, v)
}

// writeLE64 writes a 64-bit value in little-endian order.
func writeLE64(buf []byte, v uint64) {
	binary.LittleEndian.PutUint64(buf, v)
}
