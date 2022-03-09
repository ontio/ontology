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
