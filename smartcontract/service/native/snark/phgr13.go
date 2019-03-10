package snark

import (
	"errors"
	"math/big"

	"github.com/ontio/ontology/smartcontract/common"

	"github.com/clearmatics/bn256"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
)

const fieldSize = 32
const g1Size = 64
const g2Size = 128

func deserializeG1(point *bn256.G1, source common.ZeroCopySource) error {
	bytes, eof := source.NextBytes(g1Size)
	if eof {
		return errors.New("eof")
	}
	_, ok := point.Unmarshal(bytes)
	if !ok {
		return errors.New("failed to unmarshal G1 point")
	}
}

func deserializeG2(point *bn256.G2, source common.ZeroCopySource) error {
	bytes, eof := source.NextBytes(g2Size)
	if eof {
		return errors.New("eof")
	}
	_, ok := point.Unmarshal(bytes)
	if !ok {
		return errors.New("failed to unmarshal G1 point")
	}
}

type phgr13VerifyingKey struct {
	icLen      uint64
	a          *bn256.G2 // alphaA*G2
	b          *bn256.G1 // alphaB*G1
	c          *bn256.G2 // alphaC*G2
	gamma      *bn256.G2 // gamma*G2
	gammaBeta1 *bn256.G1
	gammaBeta2 *bn256.G2
	z          *bn256.G2
	ic         []*bn256.G1
}

func (vk *phgr13VerifyingKey) Deserialize(source common.ZeroCopySource) error {
	var err error
	var eof bool
	var ok bool

	vk.icLen, eof = source.NextUInt64()
	if eof {
		err = errors.New("")
		return err
	}

	// if source.Len() < 2*g1Size + 5*g2Size + icLen*g1Size {

	// }
	err = deserializeG2(vk.a, source)
	if err != nil {
		return err
	}
	err = deserializeG1(vk.b, source)
	if err != nil {
		return err
	}
	err = deserializeG2(vk.c, source)
	if err != nil {
		return err
	}
	err = deserializeG2(vk.gamma, source)
	if err != nil {
		return err
	}
	err = deserializeG1(vk.gammaBeta1, source)
	if err != nil {
		return err
	}
	err = deserializeG2(vk.gammaBeta2, source)
	if err != nil {
		return err
	}
	err = deserializeG2(vk.z, source)
	if err != nil {
		return err
	}

	if source.Len() < icLen*g1Size {
		return errors.New("")
	}
	this.ic = make([]*bn256.G1, icLen)
	for i := 0; i < icLen; i++ {
		err = deserializeG1(vk.ic[i], source)
		if err != nil {
			return err
		}
	}
	return nil
}

type phgr13Proof struct {
	a      *bn256.G1
	aPrime *bn256.G1
	b      *bn256.G2
	bPrime *bn256.G1
	c      *bn256.G1
	cPrime *bn256.G1
	k      *bn256.G1
	h      *bn256.G1
}

func (proof *phgr13Proof) Deserialize(source common.ZeroCopySource) error {
	var err error

	// if source.Len() < 7*g1Size + 1*g2Size {

	// }

	err = deserializeG1(proof.a, source)
	if err != nil {
		return err
	}
	err = deserializeG1(proof.aPrime, source)
	if err != nil {
		return err
	}
	err = deserializeG2(proof.b, source)
	if err != nil {
		return err
	}
	err = deserializeG1(proof.bPrime, source)
	if err != nil {
		return err
	}
	err = deserializeG1(proof.c, source)
	if err != nil {
		return err
	}
	err = deserializeG1(proof.cPrime, source)
	if err != nil {
		return err
	}
	err = deserializeG1(proof.k, source)
	if err != nil {
		return err
	}
	err = deserializeG1(proof.h, source)
	if err != nil {
		return err
	}
	return nil
}

func verify(vk *phgr13VerifyingKey, proof *phgr13Proof, input []*big.Int) bool {
	if len(input)+1 != len(vk.ic) {
		return false
	}
	vkX := vk.ic[0]
	for i := 0; i < len(input); i++ {
		vkX.Add(vkX, new(bn256.G1).ScalarMult(vk.ic[i+1], input[i]))
	}

	// p1 := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	p2 := new(bn256.G2).ScalarBaseMult(big.NewInt(1))
	if !bn256.PairingCheck([]*bn256.G1{proof.a, new(bn256.G1).Neg(proof.aPrime)},
		[]*bn256.G2{vk.a, p2}) {
		return false
	}
	if !bn256.PairingCheck([]*bn256.G1{vk.b, new(bn256.G1).Neg(proof.bPrime)},
		[]*bn256.G2{proof.b, p2}) {
		return false
	}
	if !bn256.PairingCheck([]*bn256.G1{proof.c, new(bn256.G1).Neg(proof.cPrime)},
		[]*bn256.G2{vk.c, p2}) {
		return false
	}

	vkxPlusAPlusC := new(bn256.G1).Add(vkX, proof.a)
	vkxPlusAPlusC.Add(vkxPlusAPlusC, proof.c)
	//vkxPlusAPlusC.Neg(vkxPlusAPlusC)
	if !bn256.PairingCheck([]*bn256.G1{proof.k, new(bn256.G1).Neg(vkxPlusAPlusC),
		new(bn256.G1).Neg(vk.gammaBeta1)}, []*bn256.G2{vk.gamma, vk.gammaBeta2, proof.b}) {
		return false
	}

	if !bn256.PairingCheck([]*bn256.G1{new(bn256.G1).Add(vkX, proof.a),
		new(bn256.G1).Neg(proof.h), new(bn256.G1).Neg(proof.c)},
		[]*bn256.G2{proof.b, vk.z, p2}) {
		return false
	}
	return true
}

// PHGR13Verify ...
func PHGR13Verify(ns *native.NativeService) ([]byte, error) {
	var err error
	// inputs are
	//  1. vk
	//  2. proof
	//  3. public input

	source := common.NewZeroCopySource(ns.Input)
	// deserialize vk
	vk := new(phgr13VerifyingKey)
	err = vk.Deserialize(source)
	if err != nil {
		return nil, err
	}

	// deserialize proof
	proof := new(phgr13Proof)
	err = proof.Deserialize(source)
	if err != nil {
		return nil, err
	}

	// deserialize public input
	// public input is a vector of field elements
	inputLen := source.NextUInt64(source)
	input := make([]*big.Int, inputLen)
	for i := 0; i < inputLen; i++ {
		bytes, eof := source.NextBytes(fieldSize)
		if eof {
			return nil, errors.New("encounter eof when deserialize public input")
		}
		input[i] = new(big.Int).SetBytes(bytes)
	}

	// do the actual zk-SNARKs verification
	valid := verify(vk, proof, input)
	if !valid {
		return common.BYTES_FALSE, nil
	} else {
		return common.BYTES_TRUE, nil
	}
}
