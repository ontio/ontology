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

package actor

import (
	"fmt"
	"reflect"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology-eventbus/actor"
)

var DefLedgerPid *actor.PID

type LedgerActor struct {
	props *actor.Props
}

func NewLedgerActor() *LedgerActor {
	return &LedgerActor{}
}

func (self *LedgerActor) Start() *actor.PID {
	self.props = actor.FromProducer(func() actor.Actor { return self })
	var err error
	DefLedgerPid, err = actor.SpawnNamed(self.props, "LedgerActor")
	if err != nil {
		panic(fmt.Errorf("LedgerActor SpawnNamed error:%s", err))
	}
	return DefLedgerPid
}

func (self *LedgerActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
	case *AddHeaderReq:
		self.handleAddHeaderReq(ctx, msg)
	case *AddHeadersReq:
		self.handleAddHeadersReq(ctx, msg)
	case *AddBlockReq:
		self.handleAddBlockReq(ctx, msg)
	case *GetTransactionReq:
		self.handleGetTransactionReq(ctx, msg)
	case *GetBlockByHashReq:
		self.handleGetBlockByHashReq(ctx, msg)
	case *GetBlockByHeightReq:
		self.handleGetBlockByHeightReq(ctx, msg)
	case *GetHeaderByHashReq:
		self.handleGetHeaderByHashReq(ctx, msg)
	case *GetHeaderByHeightReq:
		self.handleGetHeaderByHeightReq(ctx, msg)
	case *GetCurrentBlockHashReq:
		self.handleGetCurrentBlockHashReq(ctx, msg)
	case *GetCurrentBlockHeightReq:
		self.handleGetCurrentBlockHeightReq(ctx, msg)
	case *GetCurrentHeaderHeightReq:
		self.handleGetCurrentHeaderHeightReq(ctx, msg)
	case *GetCurrentHeaderHashReq:
		self.handleGetCurrentHeaderHashReq(ctx, msg)
	case *GetBlockHashReq:
		self.handleGetBlockHashReq(ctx, msg)
	case *IsContainBlockReq:
		self.handleIsContainBlockReq(ctx, msg)
	case *GetContractStateReq:
		self.handleGetContractStateReq(ctx, msg)
	case *GetMerkleProofReq:
		self.handleGetMerkleProofReq(ctx, msg)
	case *GetStorageItemReq:
		self.handleGetStorageItemReq(ctx, msg)
	case *GetBookkeeperStateReq:
		self.handleGetBookkeeperStateReq(ctx, msg)
	case *GetCurrentStateRootReq:
		self.handleGetCurrentStateRootReq(ctx, msg)
	case *IsContainTransactionReq:
		self.handleIsContainTransactionReq(ctx, msg)
	case *GetTransactionWithHeightReq:
		self.handleGetTransactionWithHeightReq(ctx, msg)
	case *GetBlockRootWithNewTxRootReq:
		self.handleGetBlockRootWithNewTxRootReq(ctx, msg)
	case *PreExecuteContractReq:
		self.handlePreExecuteContractReq(ctx, msg)
	case *GetEventNotifyByTxReq:
		self.handleGetEventNotifyByTx(ctx, msg)
	case *GetEventNotifyByBlockReq:
		self.handleGetEventNotifyByBlock(ctx, msg)
	default:
		log.Warnf("LedgerActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

func (self *LedgerActor) handleAddHeaderReq(ctx actor.Context, req *AddHeaderReq) {
	err := ledger.DefLedger.AddHeaders([]*types.Header{req.Header})
	if ctx.Sender() != nil {
		hash := req.Header.Hash()
		resp := &AddHeaderRsp{
			BlockHash: hash,
			Error:     err,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (self *LedgerActor) handleAddHeadersReq(ctx actor.Context, req *AddHeadersReq) {
	err := ledger.DefLedger.AddHeaders(req.Headers)
	if ctx.Sender() != nil {
		hashes := make([]common.Uint256, 0, len(req.Headers))
		for _, header := range req.Headers {
			hash := header.Hash()
			hashes = append(hashes, hash)
		}
		resp := &AddHeadersRsp{
			BlockHashes: hashes,
			Error:       err,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (self *LedgerActor) handleAddBlockReq(ctx actor.Context, req *AddBlockReq) {
	err := ledger.DefLedger.AddBlock(req.Block)
	if ctx.Sender() != nil {
		hash := req.Block.Hash()
		resp := &AddBlockRsp{
			BlockHash: hash,
			Error:     err,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (self *LedgerActor) handleGetTransactionReq(ctx actor.Context, req *GetTransactionReq) {
	tx, err := ledger.DefLedger.GetTransaction(req.TxHash)
	resp := &GetTransactionRsp{
		Error: err,
		Tx:    tx,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockByHashReq(ctx actor.Context, req *GetBlockByHashReq) {
	block, err := ledger.DefLedger.GetBlockByHash(req.BlockHash)
	resp := &GetBlockByHashRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockByHeightReq(ctx actor.Context, req *GetBlockByHeightReq) {
	block, err := ledger.DefLedger.GetBlockByHeight(req.Height)
	resp := &GetBlockByHeightRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetHeaderByHashReq(ctx actor.Context, req *GetHeaderByHashReq) {
	header, err := ledger.DefLedger.GetHeaderByHash(req.BlockHash)
	resp := &GetHeaderByHashRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetHeaderByHeightReq(ctx actor.Context, req *GetHeaderByHeightReq) {
	header, err := ledger.DefLedger.GetHeaderByHeight(req.Height)
	resp := &GetHeaderByHeightRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentBlockHashReq(ctx actor.Context, req *GetCurrentBlockHashReq) {
	curBlockHash := ledger.DefLedger.GetCurrentBlockHash()
	resp := &GetCurrentBlockHashRsp{
		BlockHash: curBlockHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentBlockHeightReq(ctx actor.Context, req *GetCurrentBlockHeightReq) {
	curBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()
	resp := &GetCurrentBlockHeightRsp{
		Height: curBlockHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Sender())
}

func (self *LedgerActor) handleGetCurrentHeaderHeightReq(ctx actor.Context, req *GetCurrentHeaderHeightReq) {
	curHeaderHeight := ledger.DefLedger.GetCurrentHeaderHeight()
	resp := &GetCurrentHeaderHeightRsp{
		Height: curHeaderHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentHeaderHashReq(ctx actor.Context, req *GetCurrentHeaderHashReq) {
	curHeaderHash := ledger.DefLedger.GetCurrentHeaderHash()
	resp := &GetCurrentHeaderHashRsp{
		BlockHash: curHeaderHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockHashReq(ctx actor.Context, req *GetBlockHashReq) {
	hash := ledger.DefLedger.GetBlockHash(req.Height)
	resp := &GetBlockHashRsp{
		BlockHash: hash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleIsContainBlockReq(ctx actor.Context, req *IsContainBlockReq) {
	con, err := ledger.DefLedger.IsContainBlock(req.BlockHash)
	resp := &IsContainBlockRsp{
		IsContain: con,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetContractStateReq(ctx actor.Context, req *GetContractStateReq) {
	state, err := ledger.DefLedger.GetContractState(req.ContractHash)
	resp := &GetContractStateRsp{
		ContractState: state,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetMerkleProofReq(ctx actor.Context, req *GetMerkleProofReq) {
	state, err := ledger.DefLedger.GetMerkleProof(req.ProofHeight, req.RootHeight)
	resp := &GetMerkleProofRsp{
		Proof: state,
		Error: err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockRootWithNewTxRootReq(ctx actor.Context, req *GetBlockRootWithNewTxRootReq) {
	newRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(req.TxRoot)
	resp := &GetBlockRootWithNewTxRootRsp{
		NewTxRoot: newRoot,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetTransactionWithHeightReq(ctx actor.Context, req *GetTransactionWithHeightReq) {
	tx, height, err := ledger.DefLedger.GetTransactionWithHeight(req.TxHash)
	resp := &GetTransactionWithHeightRsp{
		Tx:     tx,
		Height: height,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentStateRootReq(ctx actor.Context, req *GetCurrentStateRootReq) {
	stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
	resp := &GetCurrentStateRootRsp{
		StateRoot: stateRoot,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBookkeeperStateReq(ctx actor.Context, req *GetBookkeeperStateReq) {
	bookKeep, err := ledger.DefLedger.GetBookkeeperState()
	resp := &GetBookkeeperStateRsp{
		BookKeepState: bookKeep,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetStorageItemReq(ctx actor.Context, req *GetStorageItemReq) {
	value, err := ledger.DefLedger.GetStorageItem(req.CodeHash, req.Key)
	resp := &GetStorageItemRsp{
		Value: value,
		Error: err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleIsContainTransactionReq(ctx actor.Context, req *IsContainTransactionReq) {
	isCon, err := ledger.DefLedger.IsContainTransaction(req.TxHash)
	resp := &IsContainTransactionRsp{
		IsContain: isCon,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handlePreExecuteContractReq(ctx actor.Context, req *PreExecuteContractReq) {
	result, err := ledger.DefLedger.PreExecuteContract(req.Tx)
	resp := &PreExecuteContractRsp{
		Result: result,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetEventNotifyByTx(ctx actor.Context, req *GetEventNotifyByTxReq) {
	result, err := ledger.DefLedger.GetEventNotifyByTx(req.Tx)
	resp := &GetEventNotifyByTxRsp{
		Notifies: result,
		Error:    err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetEventNotifyByBlock(ctx actor.Context, req *GetEventNotifyByBlockReq) {
	result, err := ledger.DefLedger.GetEventNotifyByBlock(req.Height)
	resp := &GetEventNotifyByBlockRsp{
		TxHashes: result,
		Error:    err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}
