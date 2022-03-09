// Copyright (C) 2021 The Ontology Authors
// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package witness

import (
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

const WitnessGlobalParamKey = "evm.witness" // value is the deployed hex addresss(evm form) of witness contract.

var EventWitnessedEventID = crypto.Keccak256Hash([]byte("EventWitnessed(address,bytes32)"))

type EventWitnessEvent struct {
	Sender common.Address
	Hash   common.Uint256
}

func DecodeEventWitness(log *types.StorageLog) (*EventWitnessEvent, error) {
	if len(log.Topics) != 3 {
		return nil, errors.New("witness: wrong topic number")
	}
	if log.Topics[0] != EventWitnessedEventID {
		return nil, errors.New("witness: wrong event id")
	}

	sender, err := common.AddressParseFromBytes(log.Topics[1][12:])
	if err != nil {
		return nil, err
	}

	return &EventWitnessEvent{
		Sender: sender,
		Hash:   common.Uint256(log.Topics[2]),
	}, nil
}
