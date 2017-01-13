package vm

import (
	"fmt"
)

func ExampleVmReader() {
	vr := NewVmReader( []byte{0x1,0x2,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xA,0xB,0xC,0xD,0xE,0xF,0x10,0x11,0x12,0x13} )
	vr1 := NewVmReader( []byte{0x14,0x15,0x16,0x17,0x18,0x19,0x1,0x2,0x3,0x4,0x5,0x6,0x7,0x8,0x9,0xa} )

	fmt.Println( "NewVmReader() test:", vr.BaseStream )

	bt := vr.ReadByte()
	fmt.Println( "ReadByte() test:", bt,vr.ReadByte() )

	bb := vr.ReadBytes(3)
	fmt.Println( "ReadBytes() test:", bb )

	fmt.Println( "ReadUint16() test:", vr.ReadUint16() )

	fmt.Println( "ReadUInt32() test:", vr.ReadUInt32() )

	fmt.Println( "ReadUInt64() test:", vr.ReadUInt64() )

	fmt.Println( "ReadInt16() test:", vr1.ReadInt16() )

	fmt.Println( "ReadInt32() test:", vr1.ReadInt32() )

	fmt.Println( "Position() test:", vr1.Position() )

	fmt.Println( "Length() test:", vr1.Length() )

	offset, _ := vr1.Seek( 1, 1 )
	fmt.Println( "Seek() test:", offset )

	//bb1 := vr1.ReadVarInt( 999 )
	//fmt.Println( "ReadVarInt() test:", bb1 )


	// output: ok
}