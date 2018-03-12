package actor

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	"reflect"
)

var DefLedgerPid *actor.PID

type LedgerActor struct {
	props *actor.Props
}

func NewLedgerActor() *LedgerActor {
	return &LedgerActor{}
}

func (this *LedgerActor) Start() *actor.PID {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	DefLedgerPid = actor.Spawn(this.props)
	return DefLedgerPid
}

func (this *LedgerActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *AddHeaderReq:
		this.handleAddHeaderReq(ctx, msg)
	case *AddHeadersReq:
		this.handleAddHeadersReq(ctx, msg)
	case *AddBlockReq:
		this.handleAddBlockReq(ctx, msg)
	case *GetTransactionReq:
		this.handleGetTransactionReq(ctx, msg)
	case *GetBlockByHashReq:
		this.handleGetBlockByHashReq(ctx, msg)
	case *GetBlockByHeightReq:
		this.handleGetBlockByHeightReq(ctx, msg)
	case *GetHeaderByHashReq:
		this.handleGetHeaderByHashReq(ctx, msg)
	case *GetHeaderByHeightReq:
		this.handleGetHeaderByHeightReq(ctx, msg)
	case *GetCurrentBlockHashReq:
		this.handleGetCurrentBlockHashReq(ctx, msg)
	case *GetCurrentBlockHeightReq:
		this.handleGetCurrentBlockHeightReq(ctx, msg)
	case *GetCurrentHeaderHeightReq:
		this.handleGetCurrentHeaderHeightReq(ctx, msg)
	case *GetBlockHashReq:
		this.handleGetBlockHashReq(ctx, msg)
	case *IsContainBlockReq:
		this.handleIsContainBlockReq(ctx, msg)
	case *GetContractStateReq:
		this.handleGetContractStateReq(ctx, msg)
	case *GetStorageItemReq:
		this.handleGetStorageItemReq(ctx, msg)
	case *GetBookKeeperStateReq:
		this.handleGetBookKeeperStateReq(ctx, msg)
	case *GetCurrentStateRootReq:
		this.handleGetCurrentStateRootReq(ctx, msg)
	case *IsContainTransactionReq:
		this.handleIsContainTransactionReq(ctx, msg)
	case *GetTransactionWithHeightReq:
		this.handleGetTransactionWithHeightReq(ctx, msg)
	case *GetBlockRootWithNewTxRootReq:
		this.handleGetBlockRootWithNewTxRootReq(ctx, msg)
	case *PreExecuteContractReq:
		this.handlePreExecuteContractReq(ctx, msg)
	default:
		log.Warnf("LedgerActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

func (this *LedgerActor) handleAddHeaderReq(ctx actor.Context, req *AddHeaderReq) {
	err := ledger.DefLedger.AddHeaders([]*types.Header{req.Header})
	hash := req.Header.Hash()
	resp := &AddHeaderRsp{
		BlockHash: &hash,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleAddHeadersReq(ctx actor.Context, req *AddHeadersReq) {
	err := ledger.DefLedger.AddHeaders(req.Headers)
	hashes := make([]*common.Uint256, 0, len(req.Headers))
	for _, header := range req.Headers {
		hash := header.Hash()
		hashes = append(hashes, &hash)
	}
	resp := &AddHeadersRsp{
		BlockHashes: hashes,
		Error:       err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleAddBlockReq(ctx actor.Context, req *AddBlockReq) {
	err := ledger.DefLedger.AddBlock(req.Block)
	hash := req.Block.Hash()
	resp := &AddBlockRsp{
		BlockHash: &hash,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetTransactionReq(ctx actor.Context, req *GetTransactionReq) {
	tx, err := ledger.DefLedger.GetTransaction(req.TxHash)
	resp := GetTransactionRsp{
		Error: err,
		Tx:    tx,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockByHashReq(ctx actor.Context, req *GetBlockByHashReq) {
	block, err := ledger.DefLedger.GetBlockByHash(req.BlockHash)
	resp := &GetBlockByHashRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockByHeightReq(ctx actor.Context, req *GetBlockByHeightReq) {
	block, err := ledger.DefLedger.GetBlockByHeight(req.Height)
	resp := &GetBlockByHeightRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetHeaderByHashReq(ctx actor.Context, req *GetHeaderByHashReq) {
	header, err := ledger.DefLedger.GetHeaderByHash(req.BlockHash)
	resp := &GetHeaderByHashRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetHeaderByHeightReq(ctx actor.Context, req *GetHeaderByHeightReq) {
	header, err := ledger.DefLedger.GetHeaderByHeight(req.Height)
	resp := &GetHeaderByHeightRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentBlockHashReq(ctx actor.Context, req *GetCurrentBlockHashReq) {
	curBlockHash := ledger.DefLedger.GetCurrentBlockHash()
	resp := &GetCurrentBlockHashRsp{
		BlockHash: curBlockHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentBlockHeightReq(ctx actor.Context, req *GetCurrentBlockHeightReq) {
	curBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()
	resp := &GetCurrentBlockHeightRsp{
		Height: curBlockHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Sender())
}

func (this *LedgerActor) handleGetCurrentHeaderHeightReq(ctx actor.Context, req *GetCurrentHeaderHeightReq) {
	curHeaderHeight := ledger.DefLedger.GetCurrentHeaderHeight()
	resp := &GetCurrentHeaderHeightRsp{
		Height: curHeaderHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockHashReq(ctx actor.Context, req *GetBlockHashReq) {
	hash := ledger.DefLedger.GetBlockHash(req.Height)
	resp := &GetBlockHashRsp{
		BlockHash: hash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleIsContainBlockReq(ctx actor.Context, req *IsContainBlockReq) {
	con, err := ledger.DefLedger.IsContainBlock(req.BlockHash)
	resp := &IsContainBlockRsp{
		IsContain: con,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetContractStateReq(ctx actor.Context, req *GetContractStateReq) {
	state, err := ledger.DefLedger.GetContractState(req.ContractHash)
	resp := &GetContractStateRsp{
		ContractState: state,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockRootWithNewTxRootReq(ctx actor.Context, req *GetBlockRootWithNewTxRootReq) {
	newRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(req.TxRoot)
	resp := &GetBlockRootWithNewTxRootRsp{
		NewTxRoot: newRoot,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetTransactionWithHeightReq(ctx actor.Context, req *GetTransactionWithHeightReq) {
	tx, height, err := ledger.DefLedger.GetTransactionWithHeight(req.TxHash)
	resp := &GetTransactionWithHeightRsp{
		Tx:     tx,
		Height: height,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentStateRootReq(ctx actor.Context, req *GetCurrentStateRootReq) {
	stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
	resp := &GetCurrentStateRootRsp{
		StateRoot: stateRoot,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBookKeeperStateReq(ctx actor.Context, req *GetBookKeeperStateReq) {
	bookKeep, err := ledger.DefLedger.GetBookKeeperState()
	resp := &GetBookKeeperStateRsp{
		BookKeepState: bookKeep,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetStorageItemReq(ctx actor.Context, req *GetStorageItemReq) {
	value, err := ledger.DefLedger.GetStorageItem(req.CodeHash, req.Key)
	resp := &GetStorageItemRsp{
		Value: value,
		Error: err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleIsContainTransactionReq(ctx actor.Context, req *IsContainTransactionReq) {
	isCon, err := ledger.DefLedger.IsContainTransaction(req.TxHash)
	resp := &IsContainTransactionRsp{
		IsContain: isCon,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handlePreExecuteContractReq(ctx actor.Context, req *PreExecuteContractReq) {
	result, err := ledger.DefLedger.PreExecuteContract(req.Tx)
	resp := PreExecuteContractRsp{
		Result: result,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}
