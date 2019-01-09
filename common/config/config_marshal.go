package config

import (
	"io"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"encoding/json"
)

func (this *GenesisConfig) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.SeedList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize seedlist error!")
	}
	for _, s := range this.SeedList {
		if err := serialization.WriteString(w, s); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize seed error!")
		}
	}
	if err := serialization.WriteString(w, this.ConsensusType); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize consensus type error!")
	}

	switch this.ConsensusType {
	case CONSENSUS_TYPE_VBFT:
		return this.VBFT.Serialize(w)
	case CONSENSUS_TYPE_DBFT:
		return this.DBFT.Serialize(w)
	case CONSENSUS_TYPE_SOLO:
		return this.SOLO.Serialize(w)
	}
	return nil
}

func (this *GenesisConfig) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize seedlist error!")
	}
	seedlist := make([]string, 0)
	for i := 0; i < int(n); i++ {
		seed, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize seed error!")
		}
		seedlist = append(seedlist, seed)
	}

	consensusType, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize consensus type error!")
	}

	switch consensusType {
	case CONSENSUS_TYPE_VBFT:
		vbft := new(VBFTConfig)
		if err := vbft.Deserialize(r); err != nil {
			return err
		}
		this.VBFT = vbft
	case CONSENSUS_TYPE_DBFT:
		dbft := new(DBFTConfig)
		if err := dbft.Deserialize(r); err != nil {
			return err
		}
		this.DBFT = dbft
	case CONSENSUS_TYPE_SOLO:
		solo := new(SOLOConfig)
		if err := solo.Deserialize(r); err != nil {
			return err
		}
		this.SOLO = solo
	}

	this.SeedList = seedlist
	this.ConsensusType = consensusType
	return nil
}

func (this *DBFTConfig) Serialize(w io.Writer) error {
	return jsonSerialize(w, this)
}

func (this *DBFTConfig) Deserialize(r io.Reader) error {
	return jsonDeserialize(r, this)
}

func (this *SOLOConfig) Serialize(w io.Writer) error {
	return jsonSerialize(w, this)
}

func (this *SOLOConfig) Deserialize(r io.Reader) error {
	return jsonDeserialize(r, this)
}

//
// Note:
// only serialize (genesis, common, consensus, p2pnode, shard),
// (rpc, restful, ws) is not included
//
func (this *OntologyConfig) Serialize(w io.Writer) error {
	if err := this.Genesis.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "OntologyConfig serialization, serialize genesis error!")
	}

	if err := jsonSerialize(w, this.Common); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "OntologyConfig serialization, serialize common error!")
	}
	if err := jsonSerialize(w, this.Consensus); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "OntologyConfig serialization, serialize common error!")
	}
	if err := jsonSerialize(w, this.P2PNode); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "OntologyConfig serialization, serialize common error!")
	}
	if err := jsonSerialize(w, this.Shard); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "OntologyConfig serialization, serialize common error!")
	}

	return nil
}

//
// Note:
// only deserialize (genesis, common, consensus, p2pnode, shard),
// (rpc, restful, ws) is not included
//
func (this *OntologyConfig) Deserialize(r io.Reader) error {
	genesis := new(GenesisConfig)
	if err := genesis.Deserialize(r); err != nil {
		return err
	}

	common := new(CommonConfig)
	if err := jsonDeserialize(r, common); err != nil {
		return err
	}

	consensus := new(ConsensusConfig)
	if err := jsonDeserialize(r, consensus); err != nil {
		return err
	}

	p2pnode := new(P2PNodeConfig)
	if err := jsonDeserialize(r, p2pnode); err != nil {
		return err
	}

	shard := new(ShardConfig)
	if err := jsonDeserialize(r, shard); err != nil {
		return err
	}

	this.Genesis = genesis
	this.Common = common
	this.Consensus = consensus
	this.P2PNode = p2pnode
	this.Shard = shard

	// disable rpc, restful and websocket
	this.Rpc = &RpcConfig{ EnableHttpJsonRpc: false }
	this.Restful = &RestfulConfig{ EnableHttpRestful: false }
	this.Ws = &WebSocketConfig{ EnableHttpWs: false }
	return nil
}

func jsonSerialize(w io.Writer, v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "jsonSerialize")
	}

	if err := serialization.WriteVarBytes(w, buf); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarBytes, serialize buf len error!")
	}
	return nil
}

func jsonDeserialize(r io.Reader, v interface{}) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialization.WriteVarBytes, buf len error!")
	}

	if err := json.Unmarshal(buf, v); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "json.Unmarshal error!")
	}
	return nil
}
