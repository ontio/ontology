package trie

import (
	"testing"
	"bytes"
)

func TestHexCompact(t *testing.T) {
	tests := []struct{ hex, compact []byte }{
		// empty keys, with and without terminator.
		{hex: []byte{}, compact: []byte{0x00}},
		{hex: []byte{16}, compact: []byte{0x20}},
		// odd length, no terminator
		{hex: []byte{1, 2, 3, 4, 5}, compact: []byte{0x11, 0x23, 0x45}},
		// even length, no terminator
		{hex: []byte{0, 1, 2, 3, 4, 5}, compact: []byte{0x00, 0x01, 0x23, 0x45}},
		// odd length, terminator
		{hex: []byte{15, 1, 12, 11, 8, 16}, compact: []byte{0x3f, 0x1c, 0xb8}},
		// even length, terminator
		{hex: []byte{0, 15, 1, 12, 11, 8, 16}, compact: []byte{0x20, 0x0f, 0x1c, 0xb8}},
	}
	for _, test := range tests {
		if c := hexToCompact(test.hex); !bytes.Equal(c, test.compact) {
			t.Errorf("hexToCompact(%x) -> %x, want %x", test.hex, c, test.compact)
		}
		if h := compactToHex(test.compact); !bytes.Equal(h, test.hex) {
			t.Errorf("compactToHex(%x) -> %x, want %x", test.compact, h, test.hex)
		}
	}
}

func TestHexKeybytes(t *testing.T) {
	tests := []struct{ key, hexIn, hexOut []byte }{
		{key: []byte{}, hexIn: []byte{16}, hexOut: []byte{16}},
		{key: []byte{}, hexIn: []byte{}, hexOut: []byte{16}},
		{
			key:    []byte{0x12, 0x34, 0x56},
			hexIn:  []byte{1, 2, 3, 4, 5, 6, 16},
			hexOut: []byte{1, 2, 3, 4, 5, 6, 16},
		},
		{
			key:    []byte{0x12, 0x34, 0x5},
			hexIn:  []byte{1, 2, 3, 4, 0, 5, 16},
			hexOut: []byte{1, 2, 3, 4, 0, 5, 16},
		},
		{
			key:    []byte{0x12, 0x34, 0x56},
			hexIn:  []byte{1, 2, 3, 4, 5, 6},
			hexOut: []byte{1, 2, 3, 4, 5, 6, 16},
		},
	}
	for _, test := range tests {
		if h := keyBytesToHex(test.key); !bytes.Equal(h, test.hexOut) {
			t.Errorf("keybytesToHex(%x) -> %x, want %x", test.key, h, test.hexOut)
		}
		if k := keyBytesToHex(test.hexIn); !bytes.Equal(k, test.key) {
			t.Errorf("hexToKeybytes(%x) -> %x, want %x", test.hexIn, k, test.key)
		}
	}
}
