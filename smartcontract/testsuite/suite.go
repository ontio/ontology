package testsuite

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/storage"
	"io"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	utils2 "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native"
	shardsysmsg "github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
)

func RandomAddress() common.Address {
	var addr common.Address
	_, _ = rand.Read(addr[:])

	return addr
}

func InstallNativeContract(addr common.Address, actions map[string]native.Handler) {
	contract := func(native *native.NativeService) {
		for name, fun := range actions {
			native.Register(name, fun)
		}
	}
	native.Contracts[addr] = contract
}

// runtime api
func RemoteNotify(ctx *native.NativeService, param shardsysmsg.NotifyReqParam) {
	txState := ctx.MainShardTxState
	// send with minimal gas fee
	msg := &xshard_types.XShardNotify{
		NotifyID: txState.NumNotifies,
		Contract: param.ToContract,
		Payer:    ctx.Tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Method:   param.Method,
		Args:     param.Args,
	}
	txState.NumNotifies += 1
	// todo: clean shardnotifies when replay transaction
	txState.ShardNotifies = append(txState.ShardNotifies, msg)
}

// runtime api
func RemoteInvoke(ctx *native.NativeService, reqParam shardsysmsg.NotifyReqParam) ([]byte, error) {
	txState := ctx.MainShardTxState
	reqIdx := txState.NextReqID
	if reqIdx >= xshard_state.MaxRemoteReqPerTx {
		return utils.BYTE_FALSE, xshard_state.ErrTooMuchRemoteReq
	}
	msg := &xshard_types.XShardTxReq{
		IdxInTx:  uint64(reqIdx),
		Payer:    ctx.Tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Contract: reqParam.ToContract,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}
	txState.NextReqID += 1

	if reqIdx < uint32(len(txState.OutReqResp)) {
		if xshard_types.IsXShardMsgEqual(msg, txState.OutReqResp[reqIdx].Req) == false {
			return utils.BYTE_FALSE, xshard_state.ErrMismatchedRequest
		}
		rspMsg := txState.OutReqResp[reqIdx].Resp
		var resultErr error
		if rspMsg.Error {
			resultErr = errors.New("remote invoke got error response")
		}
		return rspMsg.Result, resultErr
	}

	if len(txState.TxPayload) == 0 {
		txPayload := bytes.NewBuffer(nil)
		if err := ctx.Tx.Serialize(txPayload); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to get tx payload: %s", err)
		}
		txState.TxPayload = txPayload.Bytes()
	}

	// no response found in tx-statedb, send request
	if err := txState.AddTxShard(reqParam.ToShard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("remote invoke, failed to add shard: %s", err)
	}

	txState.PendingReq = &xshard_state.XShardReqMsg{
		SourceShardID: ctx.ShardID,
		SourceHeight:  ctx.Height,
		TargetShardID: reqParam.ToShard,
		SourceTxHash:  ctx.Tx.Hash(),
		Req:           msg,
	}
	txState.ExecuteState = xshard_state.ExecYielded

	// put Tx-Request
	// todo: clean
	//if err := remoteSendShardMsg(ctx, txHash, reqParam.ToShard, msg); err != nil {
	//	return utils.BYTE_FALSE, fmt.Errorf("remote invoke, notify: %s", err)
	//}

	return utils.BYTE_FALSE, shardsysmsg.ErrYield
}

func processTransaction(blockHeight uint32, tx *types.Transaction) (result interface{}, err error) {
	overlay := NewOverlayDB()
	xshardDB := storage.NewXShardDB(overlay)
	cache := storage.NewCacheDB(overlay)

	txState, result, err := executeTransaction(tx, cache)
	if err != nil {
		if txState != nil && txState.PendingReq != nil { // yielded
			for id := range txState.Shards {
				xshardDB.AddToShard(blockHeight, id)
			}
			_ = xshardDB.AddXShardReqsInBlock(blockHeight, &xshard_types.CommonShardMsg{
				SourceTxHash:  txState.PendingReq.SourceTxHash,
				SourceShardID: txState.PendingReq.SourceShardID,
				SourceHeight:  uint64(txState.PendingReq.SourceHeight),
				TargetShardID: txState.PendingReq.TargetShardID,
				Msg:           txState.PendingReq.Req,
			})

			//xshardDB.SetShardTxState()
		}
	}

	return result, err
}

func executeTransaction(tx *types.Transaction, cache *storage.CacheDB) (*xshard_state.TxState,
	interface{}, error) {
	config := &smartcontract.Config{
		ShardID: common.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	txHash := tx.Hash()
	txState := xshard_state.CreateTxState(xshard_state.ShardTxID(string(txHash[:])))

	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
			PreExec:          true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingReq != nil {
				return txState, nil, shardsysmsg.ErrYield
			}
			return nil, nil, err
		}

		return txState, res, nil
	}

	panic("unimplemented")
}

func resumeTx(shardTxID xshard_state.ShardTxID, rspMsg *xshard_types.XShardTxRsp) (*xshard_state.TxState, interface{}, error) {
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	if txState.PendingReq == nil || txState.PendingReq.Req.IdxInTx != rspMsg.IdxInTx {
		// todo: system error or remote shard error
		return nil, nil, fmt.Errorf("invalid response id: %d", rspMsg.IdxInTx)
	}

	txState.OutReqResp = append(txState.OutReqResp, &xshard_state.XShardTxReqResp{Req: txState.PendingReq.Req, Resp: rspMsg})
	txState.PendingReq = nil

	txPayload := txState.TxPayload
	if txPayload == nil {
		return nil, nil, fmt.Errorf("failed to get tx payload")
	}

	// FIXME: invoke neo contract
	// re-execute tx
	txState.NextReqID = 0

	tx, err := types.TransactionFromRawBytes(txPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to re-init original tx: %s", err)
	}

	config := &smartcontract.Config{
		ShardID: common.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)
	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
			PreExec:          true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingReq != nil {
				return txState, nil, shardsysmsg.ErrYield
			}
			return nil, nil, err
		}

		// xshard transaction has completed
		txState.WriteSet = cache.GetCache()

		return txState, res, nil
	}

	panic("unimplemented")
}

func processXShardRsp(shardTxID xshard_state.ShardTxID, rspMsg *xshard_types.XShardTxRsp) (*xshard_state.TxState, error) {
	txState, _, err := resumeTx(shardTxID, rspMsg)

	if err != nil {
		if err == shardsysmsg.ErrYield {
			return txState, err
		}
		// Txn failed, abort all transactions
		//todo abort tx
		//if _, err2 := abortTx(ctx, txState, tx); err2 != nil {
		//	return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		//}
		//return resultErr
		return txState, err
	}

	return txState, err
}

var ShardA = common.NewShardIDUnchecked(1)
var ShardB = common.NewShardIDUnchecked(2)

func RemoteNotifyPing(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteString(fmt.Sprintf("hello from shard: %d", native.ShardID.ToUint64()))

	params := shardsysmsg.NotifyReqParam{
		ToShard: ShardB,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	RemoteNotify(native, params)

	return utils.BYTE_TRUE, nil
}

func RemoteInvokeAddAndInc(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)

	params := shardsysmsg.NotifyReqParam{
		ToShard: ShardB,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	sum, err := RemoteInvoke(native, params)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(sum)
	s, eof := source.NextUint64()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}

	sink.Reset()
	sink.WriteUint64(s + 1)

	return sink.Bytes(), err
}

func HandlePing(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_TRUE, nil
}

func BuildInvokeTx(contractAddress common.Address, method string,
	args []interface{}) *types.Transaction {
	invokCode, err := utils2.BuildNativeInvokeCode(contractAddress, 0, method, args)
	if err != nil {
		return nil
	}
	invokePayload := &payload.InvokeCode{
		Code: invokCode,
	}
	tx := &types.MutableTransaction{
		Version:  0,
		GasPrice: 0,
		GasLimit: 1000000000,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	res, err := tx.IntoImmutable()
	if err != nil {
		return nil
	}
	return res
}
