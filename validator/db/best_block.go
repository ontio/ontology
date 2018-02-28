package db

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
)

type BestBlock struct {
	Height uint32
	Hash   common.Uint256
}

type BestStateProvider interface {
	GetBestBlock() (BestBlock, error)
	GetBestHeader() (*types.Header, error)
}
