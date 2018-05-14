package ontid

import (
	"encoding/hex"
	"errors"

	"github.com/ontio/ontology-crypto/keypair"
	cmn "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
)

const flag_exist = 0x01

func checkIDExistence(srvc *native.NativeService, encID []byte) bool {
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, encID)
	if err == nil {
		t, ok := val.(*states.StorageItem)
		if ok {
			if len(t.Value) > 0 && t.Value[0] == flag_exist {
				return true
			}
		}
	}
	return false
}

const (
	FIELD_PK byte = 1 + iota
	FIELD_PK_STATE
	FIELD_ATTR
	FIELD_RECOVERY
)

func encodeID(id []byte) ([]byte, error) {
	length := len(id)
	if length == 0 || length > 255 {
		return nil, errors.New("encode ONT ID error: invalid ID length")
	}
	enc := []byte{byte(length)}
	enc = append(enc, id...)
	enc = append(genesis.OntIDContractAddress[:], enc...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) != int(data[0])+1 {
		return nil, errors.New("decode ONT ID error: invalid data length")
	}
	return data[1:], nil
}

func setRecovery(srvc *native.NativeService, encID, recovery []byte) error {
	key := append(encID, FIELD_RECOVERY)
	val := &states.StorageItem{Value: recovery}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := getStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	}
	return item.Value, nil
}

func getStorageItem(srvc *native.NativeService, key []byte) (*states.StorageItem, error) {
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, key)
	if err != nil {
		return nil, err
	}
	t, ok := val.(*states.StorageItem)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	return t, nil
}

func checkWitness(srvc *native.NativeService, key []byte) error {
	// try as if key is an address
	addr, err := cmn.AddressParseFromBytes(key)
	if srvc.ContextRef.CheckWitness(addr) {
		return nil
	}

	// try as if key is a public key
	pk, err := keypair.DeserializePublicKey(key)
	if err != nil {
		return errors.New("invalid public key, " + err.Error())
	}
	addr = types.AddressFromPubKey(pk)
	if srvc.ContextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}
