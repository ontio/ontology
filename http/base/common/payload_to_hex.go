package common

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/ontio/ontology-crypto/keypair"
)

type PayloadInfo interface{}

//implement PayloadInfo define BookKeepingInfo
type BookKeepingInfo struct {
	Nonce  uint64
}

type InvokeCodeInfo struct {
	Code     string
	GasLimit uint64
	VmType   int
}
type DeployCodeInfo struct {
	VmType      int
	Code        string
	NeedStorage bool
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
}

//implement PayloadInfo define IssueAssetInfo
type IssueAssetInfo struct {
}


//implement PayloadInfo define TransferAssetInfo
type TransferAssetInfo struct {
}

type RecordInfo struct {
	RecordType string
	RecordData string
}

type BookkeeperInfo struct {
	PubKey     string
	Action     string
	Issuer     string
	Controller string
}

type DataFileInfo struct {
	IPFSPath string
	Filename string
	Note     string
	Issuer   string
}

type Claim struct {
	Claims []*UTXOTxInput
}

type UTXOTxInput struct {
	ReferTxID          string
	ReferTxOutputIndex uint16
}

type PrivacyPayloadInfo struct {
	PayloadType uint8
	Payload     string
	EncryptType uint8
	EncryptAttr string
}

type VoteInfo struct {
	PubKeys []string
	Voter   string
}

func TransPayloadToHex(p types.Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.BookKeeping:
		obj := new(BookKeepingInfo)
		obj.Nonce = object.Nonce
		return obj
	case *payload.Bookkeeper:
		obj := new(BookkeeperInfo)
		pubKeyBytes := keypair.SerializePublicKey(object.PubKey)
		obj.PubKey = ToHexString(pubKeyBytes)
		if object.Action == payload.BookKeeperAction_ADD {
			obj.Action = "add"
		} else if object.Action == payload.BookKeeperAction_SUB {
			obj.Action = "sub"
		} else {
			obj.Action = "nil"
		}
		pubKeyBytes = keypair.SerializePublicKey(object.Issuer)
		obj.Issuer = ToHexString(pubKeyBytes)

		return obj
	case *payload.InvokeCode:
		obj := new(InvokeCodeInfo)
		obj.Code = ToHexString(object.Code.Code)
		obj.GasLimit = uint64(object.GasLimit)
		obj.VmType = int(object.Code.VmType)
		return obj
	case *payload.DeployCode:
		obj := new(DeployCodeInfo)
		obj.VmType = int(object.Code.VmType)
		obj.Code = ToHexString(object.Code.Code)
		obj.NeedStorage = object.NeedStorage
		obj.Name = object.Name
		obj.CodeVersion = object.Version
		obj.Author = object.Author
		obj.Email = object.Email
		obj.Description = object.Description
		return obj
	case *payload.Vote:
		obj := new(VoteInfo)
		obj.PubKeys = make([]string, len(object.PubKeys))
		obj.Voter = ToHexString(object.Account[:])
		for i, key := range object.PubKeys {
			pubKeyBytes := keypair.SerializePublicKey(key)
			obj.PubKeys[i] = ToHexString(pubKeyBytes)
		}
	}
	return nil
}
