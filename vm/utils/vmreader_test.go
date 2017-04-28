package utils

import (
	"testing"
)

func TestExampleVmReader(t *testing.T) {
	vr := NewVmReader( []byte{0x1,0x2,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	vr1 := NewVmReader( []byte{0x14,0x15,0x16,0x17,0x18,0x19,0x1,0x2,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xa} )

	t.Log( "NewVmReader() test:", vr )
	t.Log( "NewVmReader() test:", vr.BaseStream )

	//bt := vr.ReadByte()
	//t.Log( "ReadByte() test:", bt,vr.ReadByte() )

	bb := vr.ReadBytes(4)
	t.Log( "ReadBytes() test:", bb )

	t.Log( "ReadUint16() test:", vr.ReadUint16() )

	t.Log( "ReadUInt32() test:", vr.ReadUInt32() )

	t.Log( "ReadUInt64() test:", vr.ReadUInt64() )

	t.Log( "ReadInt16() test:", vr1.ReadInt16() )

	t.Log( "ReadInt32() test:", vr1.ReadInt32() )

	t.Log( "Position() test:", vr1.Position() )

	t.Log( "Length() test:", vr1.Length() )

	offset, _ := vr1.Seek( 1, 1 )
	t.Log( "Seek() test:", offset )

	//bb1 := vr1.ReadVarInt( 999 )
	//t.Log( "ReadVarInt() test:", bb1 )


	// output: ok
}
