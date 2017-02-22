package common

import (
	"bytes"
	"encoding/binary"
	_ "io"

	 "golang.org/x/crypto/ripemd160"
	"crypto/sha256"
	. "GoOnchain/errors"
	"errors"
	"io"
)

func ToCodeHash(code []byte) (Uint160,error){
	//TODO: ToCodeHash
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	io.WriteString(md, string(temp[:]))
	f := md.Sum(nil)

	hash,err := Uint160ParseFromBytes(f)
	if err != nil{
		return Uint160{},NewDetailErr(errors.New("[Common] , ToCodeHash err."), ErrNoCode, "");
	}
	return hash,nil
}

func GetNonce() uint64 {
	//TODO: GetNonce()
	return 0
}

func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt16(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int16
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int16(tmp)
}

func IsEqualBytes(b1 []byte, b2 []byte) bool {
	len1 := len(b1)
	len2 := len(b2)
	if len1 != len2 {return false}

	for i:=0; i<len1; i++ {
		if b1[i] != b2[i] {return false}
	}

	return true
}

func ToHexString(data []byte) string {
	//TODO: ToHexString
	return string(data)
}

func HexToBytes(value string) []byte {
	//TODO: HexToBytes
	return nil
}
