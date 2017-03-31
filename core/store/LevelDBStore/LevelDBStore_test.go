package LevelDBStore

import (
	"fmt"
	"bytes"
	"encoding/binary"
	"testing"
	. "DNA/common"
)

var ldbs * LevelDBStore

func byteLittleEndian( inarray []byte ) ([]byte){
	l := len(inarray)
	var out []byte = make([]byte, l)

	for  i:=0; i<l; i++ {
		out[i] = inarray[l-1-i]
	}

	return out
}

func TestNewLevelDBStore( t *testing.T ) {

	ldbs, _ = NewLevelDBStore("E:\\AntSharesCore\\Chain")

	ldbs.Put( []byte("a"), []byte("123") )

	fmt.Println( ldbs.Get( []byte("a") ) )

	ldbs.Delete( []byte("a") )

	fmt.Println( ldbs.Get( []byte("a") ) )

	ldbs.Put( []byte("a"), []byte("123") )
	ldbs.Put( []byte("b"), []byte("456") )
	ldbs.Put( []byte("c"), []byte("789") )
/*
	ro := opt.ReadOptions{
	}

	ldbs.it = ldbs.NewIterator(&ro)
	for ldbs.it.Next() {
		fmt.Println(ldbs.it.Key())
		fmt.Println(ldbs.it.Value())
	}
*/
	// CFG_Version
	b,_ := ldbs.Get( []byte("\xf0") )
	fmt.Println( "CFG_Version:", string(b) )

	// SYS_CurrentBlock
	cb,_ := ldbs.Get( []byte("\x40") )

	cbf := byteLittleEndian( cb[0:32] )
	fmt.Printf( "BigEndian Hash: %x\n", cb[0:32] )
	fmt.Printf( "SYS_CurrentBlock Hash: %x\n", cbf )

	buf := bytes.NewReader(cb[32:])
	var nHeight uint32
	binary.Read(buf, binary.LittleEndian, &nHeight )
	fmt.Println( "SYS_CurrentBlock Height:", nHeight )

	// SYS_CurrentHeader
	ch,_ := ldbs.Get( []byte("\x41") )

	chf := byteLittleEndian( ch[0:32] )
	fmt.Printf( "BigEndian Hash: %x\n", ch[0:32] )
	fmt.Printf( "SYS_CurrentHeader Hash: %x\n", chf )

	buf = bytes.NewReader(ch[32:])
	binary.Read(buf, binary.LittleEndian, &nHeight )
	fmt.Println( "SYS_CurrentHeader Height:", nHeight )

	// DATA_Header
	dh,_ := ldbs.Get( []byte("\x01\xbf\x44\x21\xc8\x87\x76\xc5\x3b\x43\xce\x1d\xc4\x54\x63\xbf\xd2\x02\x8e\x32\x2f\xdf\xb6\x00\x64\xbe\x15\x0e\xd3\xe3\x61\x25\xd4") )
	fmt.Printf( "DATA_Header: %x\n", dh )

	// DATA_Header
	//dh,_ := ldbs.Get( []byte("\x01\x8a\xe2\xf1\x10\xea\xdb\x74\xc5\xe8\xed\xa1\xd8\x38\x86\xfe\xbb\xf6\x7e\x90\xfe\x48\xc1\x77\x31\xe9\x9e\x5e\xff\xed\xdf\x24\x28") )
	//fmt.Printf( "DATA_Header: %x\n", dh )

	// DATA_Header
	//dht,_ := ldbs.Get( []byte("\x01\x28\x3a\xef\xfc\x08\x73\x3d\xf6\x95\x27\x4c\x8d\x7d\x64\x3e\x71\x20\x3d\x9e\x1e\x0d\x43\xb2\x74\x64\x78\x3d\xe6\xa3\x0a\x7c\xe6") )
	//fmt.Println( "DATA_Header: ", dht )

	// DATA_Transaction
	dt,_ := ldbs.Get( []byte("\x02\x9b\x7c\xff\xda\xa6\x74\xbe\xae\x0f\x93\x0e\xbe\x60\x85\xaf\x90\x93\xe5\xfe\x56\xb3\x4a\x5c\x22\x0c\xcd\xcf\x6e\xfc\x33\x6f\xc5") )
	fmt.Printf( "DATA_Transaction: %x\n", dt )

	// IX_HeaderHashList
	ih,_ := ldbs.Get( []byte("\x80\xD0\x07") )
	fmt.Println( ih )

	// IX_Unspent
	//iu,_ := db.Get( []byte("\x90\xfd\xa1\x49\x91\x07\x02\xcc\x19\xed\x96\x7c\x32\xf8\x83\xa3\x22\xf2\xe1\x71\x37\x90\xc1\x39\x8f\x53\x8a\x42\xe4\x89\xd4\x85\xee") )
	iu,_ := ldbs.Get( []byte("\x90\xfc\xe6\x62\xac\x7a\xd1\x45\x20\x08\xc0\x36\xaf\x92\x38\x4f\xb3\xcb\x0e\x82\xce\x4e\x8b\x50\xeb\xed\x2f\xb9\xbe\x22\xf5\xa6\xa3") )
	//iu,_ := db.Get( []byte("\x90\x7d\xed\x1c\x83\xbd\x63\xe8\x87\x1c\x8c\x2a\xd5\x76\x07\xfe\x14\x23\xe8\x79\x66\x06\xf2\xf5\xc2\xfe\x25\xbe\x3f\x27\xf8\x9a\x43") )
	fmt.Println( len(iu) )
	fmt.Println( iu )

	// ST_QuantityIssued
	st,_ := ldbs.Get( []byte("\xc1\x9b\x7c\xff\xda\xa6\x74\xbe\xae\x0f\x93\x0e\xbe\x60\x85\xaf\x90\x93\xe5\xfe\x56\xb3\x4a\x5c\x22\x0c\xcd\xcf\x6e\xfc\x33\x6f\xc5") )

	buf64 := bytes.NewReader(st)
	var asqi uint64
	binary.Read(buf64, binary.LittleEndian, &asqi )
	fmt.Println( "ST_QuantityIssued:", asqi )

	// GETBLOCK
	//bb,_ := ldbs.GetBlock( []byte("\xbf\x44\x21\xc8\x87\x76\xc5\x3b\x43\xce\x1d\xc4\x54\x63\xbf\xd2\x02\x8e\x32\x2f\xdf\xb6\x00\x64\xbe\x15\x0e\xd3\xe3\x61\x25\xd4") )
	//bb,_ := ldbs.GetBlock( []byte("\x80\x2c\xfc\xfa\x7f\x83\xe0\xc2\xda\xdb\x73\x46\x36\xc7\x58\xd3\x08\x1c\x9b\x05\x7f\xbb\x98\x14\xb3\xa8\xf4\xcb\xa3\x10\x06\xa8") )
	bb,_ := ldbs.GetBlock( Uint256{0x80,0x2c,0xfc,0xfa,0x7f,0x83,0xe0,0xc2,0xda,0xdb,0x73,0x46,0x36,0xc7,0x58,0xd3,0x08,0x1c,0x9b,0x05,0x7f,0xbb,0x98,0x14,0xb3,0xa8,0xf4,0xcb,0xa3,0x10,0x06,0xa8} )
	fmt.Printf( "Blockdata: %x\n",  bb.Blockdata )

	// GetTransaction
	//th,_ := ldbs.GetTransaction( []byte("\x9b\x7c\xff\xda\xa6\x74\xbe\xae\x0f\x93\x0e\xbe\x60\x85\xaf\x90\x93\xe5\xfe\x56\xb3\x4a\x5c\x22\x0c\xcd\xcf\x6e\xfc\x33\x6f\xc5") )
	//fmt.Printf( "Transaction: %x\n",  th )

	// SAVEBLOCK
	bb.Blockdata.Version = uint32(0xfffffffe)
	err := ldbs.SaveBlock(bb)
	fmt.Println( "SaveBlock ERR:", err )

	ldbs.Close()
}

