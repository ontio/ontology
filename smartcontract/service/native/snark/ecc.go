package snark

import (
	"errors"
	"math/big"

	"github.com/kunxian-xia/bn256"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func ECAdd(ns *native.NativeService) ([]byte, error) {
	var err error
	source := common.NewZeroCopySource(ns.Input)
	a := new(bn256.G1)
	err = deserializeG1(a, source)
	if err != nil {
		return nil, err
	}
	b := new(bn256.G1)
	err = deserializeG1(b, source)
	if err != nil {
		return nil, err
	}

	c := new(bn256.G1).Add(a, b)
	return c.Marshal(), nil
}

func TwistECAdd(ns *native.NativeService) ([]byte, error) {
	var err error
	source := common.NewZeroCopySource(ns.Input)
	a := new(bn256.G2)
	err = deserializeG2(a, source)
	if err != nil {
		return nil, err
	}
	b := new(bn256.G2)
	err = deserializeG2(b, source)
	if err != nil {
		return nil, err
	}

	c := new(bn256.G2).Add(a, b)
	return c.Marshal(), nil
}

func ECMul(ns *native.NativeService) ([]byte, error) {
	var err error
	source := common.NewZeroCopySource(ns.Input)
	p := new(bn256.G1)

	err = deserializeG1(p, source)
	if err != nil {
		return nil, err
	}
	kBytes, eof := source.NextBytes(source.Len())
	if eof {
		return nil, errors.New("read k failed: eof")
	}
	k := new(big.Int).SetBytes(kBytes)
	q := new(bn256.G1).ScalarMult(p, k)
	return q.Marshal(), nil
}

func TwistECMul(ns *native.NativeService) ([]byte, error) {
	var err error
	source := common.NewZeroCopySource(ns.Input)
	p := new(bn256.G2)

	err = deserializeG2(p, source)
	if err != nil {
		return nil, err
	}
	kBytes, eof := source.NextBytes(source.Len())
	if eof {
		return nil, errors.New("read k failed: eof")
	}
	k := new(big.Int).SetBytes(kBytes)
	q := new(bn256.G2).ScalarMult(p, k)
	return q.Marshal(), nil
}

func PairingCheck(ns *native.NativeService) ([]byte, error) {
	var err error
	if len(ns.Input)%(g1Size+g2Size) != 0 {
		return nil, errors.New("length of input is not a multiple of 192")
	}
	k := len(ns.Input) / (g1Size + g2Size)
	source := common.NewZeroCopySource(ns.Input)

	pointG1s := make([]*bn256.G1, k)
	pointG2s := make([]*bn256.G2, k)
	for i := 0; i < k; i++ {
		err = deserializeG1(pointG1s[i], source)
		if err != nil {
			return nil, err
		}
		err = deserializeG2(pointG2s[i], source)
		if err != nil {
			return nil, err
		}
	}
	if !bn256.PairingCheck(pointG1s, pointG2s) {
		return utils.BYTE_FALSE, nil
	} else {
		return utils.BYTE_TRUE, nil
	}
}
