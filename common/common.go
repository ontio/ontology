package common

import (
	. "GoOnchain/errors"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	_ "io"
	"math/rand"

	"golang.org/x/crypto/ripemd160"
	//"GoOnchain/common/log"
	"encoding/hex"
	"errors"
	"io"
)

func ToCodeHash(code []byte) (Uint160, error) {
	//TODO: ToCodeHash
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	io.WriteString(md, string(temp[:]))
	f := md.Sum(nil)

	hash, err := Uint160ParseFromBytes(f)
	if err != nil {
		return Uint160{}, NewDetailErr(errors.New("[Common] , ToCodeHash err."), ErrNoCode, "")
	}
	return hash, nil
}

func GetNonce() uint64 {
	Trace()
	// Fixme replace with the real random number generator
	nonce := uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
	Trace()
	fmt.Println(fmt.Sprintf("The new nonce is: 0x%x", nonce))
	return nonce
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
	if len1 != len2 {
		return false
	}

	for i := 0; i < len1; i++ {
		if b1[i] != b2[i] {
			return false
		}
	}

	return true
}

func ToHexString(data []byte) string {
	return hex.EncodeToString(data)
}

func HexToBytes(value string) ([]byte, error) {
	return hex.DecodeString(value)
}

func ClearBytes(arr []byte, len int) {
	for i := 0; i < len; i++ {
		arr[i] = 0
	}
}
