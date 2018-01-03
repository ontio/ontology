package types

import (
	"github.com/Ontology/common"
	"math/big"
)

func ConvertBigIntegerToBytes(data *big.Int) []byte {
	if data.Int64() == 0 {
		return []byte{}
	}

	bs := data.Bytes()
	b := bs[0]
	if data.Sign() < 0 {
		for i, b := range bs {
			bs[i] = ^b
		}
		temp := big.NewInt(0)
		temp.SetBytes(bs)
		temp2 := big.NewInt(0)
		temp2.Add(temp, big.NewInt(1))
		bs = temp2.Bytes()
		common.BytesReverse(bs)
		if b>>7 == 1 {
			bs = append(bs, 255)
		}
	} else {
		common.BytesReverse(bs)
		if b>>7 == 1 {
			bs = append(bs, 0)
		}
	}
	return bs
}

func ConvertBytesToBigInteger(ba []byte) *big.Int {
	res := big.NewInt(0)
	l := len(ba)
	if l == 0 {
		return res
	}

	bytes := make([]byte, 0, l)
	bytes = append(bytes, ba...)
	common.BytesReverse(bytes)

	if bytes[0]>>7 == 1 {
		for i, b := range bytes {
			bytes[i] = ^b
		}

		temp := big.NewInt(0)
		temp.SetBytes(bytes)
		temp2 := big.NewInt(0)
		temp2.Add(temp, big.NewInt(1))
		bytes = temp2.Bytes()
		res.SetBytes(bytes)
		return res.Neg(res)
	}

	res.SetBytes(bytes)
	return res
}
