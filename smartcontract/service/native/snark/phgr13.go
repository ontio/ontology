package snark

import (
	"errors"
	"math/big"

	"github.com/kunxian-xia/bn256"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

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

func (vk *phgr13VerifyingKey) Deserialize(source *common.ZeroCopySource) error {
	var err error
	var eof bool

	vk.icLen, eof = source.NextUint64()
	if eof {
		err = errors.New("input's length is less than required")
		return err
	}

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

	if source.Len() < vk.icLen*g1Size {
		return errors.New("input's length is less than required")
	}
	vk.ic = make([]*bn256.G1, vk.icLen)
	for i := uint64(0); i < vk.icLen; i++ {
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

func (proof *phgr13Proof) Deserialize(source *common.ZeroCopySource) error {
	var err error

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

func verify(vk *phgr13VerifyingKey, proof *phgr13Proof, input []*big.Int) (bool, error) {
	if len(input)+1 != len(vk.ic) {
		return false, errors.New("len(input) + 1 != len(vk.ic)")
	}
	vkX := vk.ic[0]
	for i := 0; i < len(input); i++ {
		vkX.Add(vkX, new(bn256.G1).ScalarMult(vk.ic[i+1], input[i]))
	}

	// p1 := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	p2 := new(bn256.G2).ScalarBaseMult(big.NewInt(1))

	if !bn256.PairingCheck([]*bn256.G1{proof.a, new(bn256.G1).Neg(proof.aPrime)},
		[]*bn256.G2{vk.a, p2}) {
		log.Error("knowledge commitments condition a failed")
		return false, nil
	}
	if !bn256.PairingCheck([]*bn256.G1{vk.b, new(bn256.G1).Neg(proof.bPrime)},
		[]*bn256.G2{proof.b, p2}) {
		log.Error("knowledge commitments condition b failed")
		return false, nil
	}
	if !bn256.PairingCheck([]*bn256.G1{proof.c, new(bn256.G1).Neg(proof.cPrime)},
		[]*bn256.G2{vk.c, p2}) {
		log.Error("knowledge commitments condition c failed")
		return false, nil
	}

	vkxPlusAPlusC := new(bn256.G1).Add(vkX, proof.a)
	vkxPlusAPlusC.Add(vkxPlusAPlusC, proof.c)

	if !bn256.PairingCheck([]*bn256.G1{proof.k, new(bn256.G1).Neg(vkxPlusAPlusC),
		new(bn256.G1).Neg(vk.gammaBeta1)}, []*bn256.G2{vk.gamma, vk.gammaBeta2, proof.b}) {
		log.Error("same coefficient condition failed")
		return false, nil
	}

	if !bn256.PairingCheck([]*bn256.G1{new(bn256.G1).Add(vkX, proof.a),
		new(bn256.G1).Neg(proof.h), new(bn256.G1).Neg(proof.c)},
		[]*bn256.G2{proof.b, vk.z, p2}) {
		log.Error("qap divisibility condition failed")
		return false, nil
	}
	return true, nil
}

// PHGR13Verify ...
func PHGR13Verify(ns *native.NativeService) ([]byte, error) {
	var err error
	// inputs consists of (vk, proof, public input)
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
	fieldsNum, eof := source.NextUint64()
	if eof {
		return nil, errors.New("eof when deserialize number of field elements in public input")
	}
	if source.Len() < fieldsNum*fieldSize {
		return nil, errors.New("input's length is less than required")
	}
	input := make([]*big.Int, fieldsNum)
	for i := uint64(0); i < fieldsNum; i++ {
		bytes, eof := source.NextBytes(fieldSize)
		if eof {
			return nil, errors.New("encounter eof when deserialize field element")
		}
		input[i] = new(big.Int).SetBytes(bytes)
	}

	// do the actual zk-SNARKs verification
	valid, err := verify(vk, proof, input)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if !valid {
		return utils.BYTE_FALSE, nil
	} else {
		return utils.BYTE_TRUE, nil
	}
}
