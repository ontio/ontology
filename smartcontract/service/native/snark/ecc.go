/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

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
