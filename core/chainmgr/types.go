package chainmgr

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
)

type shardAccount struct {
	PrivateKey []byte                    `json:"private_key"`
	PublicKey  []byte                    `json:"public_key"`
	Address    []byte                    `json:"address"`
	SigScheme  signature.SignatureScheme `json:"sig_scheme"`
}

func serializeShardAccount(acc *account.Account) ([]byte, error) {
	if acc == nil {
		return nil, fmt.Errorf("nil account")
	}

	buf := new(bytes.Buffer)
	if err := acc.Address.Serialize(buf); err != nil {
		return nil, fmt.Errorf("marshal address: %s", err)
	}

	s := &shardAccount{
		PrivateKey: keypair.SerializePrivateKey(acc.PrivKey()),
		PublicKey:  keypair.SerializePublicKey(acc.PubKey()),
		Address:    buf.Bytes(),
		SigScheme:  acc.SigScheme,
	}

	return json.Marshal(s)
}

func deserializeShardAccount(payload []byte) (*account.Account, error) {
	s := &shardAccount{}
	if err := json.Unmarshal(payload, s); err != nil {
		return nil, err
	}

	sk, err := keypair.DeserializePrivateKey(s.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("unmarshal private key: %s", err)
	}
	pk, err := keypair.DeserializePublicKey(s.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("unmarshal public key: %s", err)
	}

	var addr common.Address
	if err := addr.Deserialize(bytes.NewBuffer(s.Address)); err != nil {
		return nil, fmt.Errorf("unmarshal address: %s", err)
	}

	return &account.Account{
		PrivateKey: sk,
		PublicKey:  pk,
		Address:    addr,
		SigScheme:  s.SigScheme,
	}, nil
}

func deserializeShardConfig(payload []byte) (*config.OntologyConfig, error) {
	cfg := &config.OntologyConfig{}
	buf := bytes.NewBuffer(payload)
	if err := cfg.Deserialize(buf); err != nil {
		return nil, fmt.Errorf("deserialize ontology config: %s", err)
	}
	return cfg, nil
}
