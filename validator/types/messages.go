package types

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
)

// message
type RegisterValidator struct {
	Sender *actor.PID
	Type   VerifyType
	Id     string
}

type UnRegisterValidator struct {
	Id string
}

type UnRegisterAck struct {
	Id string
}

type CheckTx struct {
	Tx types.Transaction
}

type StatelessCheckResponse struct {
	ErrCode errors.ErrCode
	Hash    common.Uint256
}

type StatefullCheckResponse struct {
	ErrCode errors.ErrCode
	Hash    common.Uint256
	Height  int32
}

type VerifyType uint8

const (
	Stateless VerifyType = iota
	Statefull VerifyType = iota
)
