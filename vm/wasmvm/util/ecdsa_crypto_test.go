package util

import (
	"testing"
	"fmt"
)

func TestECDsaCrypto_Hash160(t *testing.T) {
	ecdsa := &ECDsaCrypto{}
	b := []byte("test string")
	res := ecdsa.Hash160(b)
	if len(res) != 20 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}
	b = []byte(nil)
	res = ecdsa.Hash160(b)
	if len(res) != 20 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}

}

func TestECDsaCrypto_Hash256(t *testing.T) {
	ecdsa := &ECDsaCrypto{}
	b := []byte("test string")
	res := ecdsa.Hash256(b)
	fmt.Println(res)
	if len(res) != 32 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}
	b = []byte(nil)
	res = ecdsa.Hash256(b)
	if len(res) != 32 {
		t.Error("TestECDsaCrypto_Hash160 length is not 20")
	}
}