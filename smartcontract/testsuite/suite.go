package testsuite

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/ontio/ontology/core/payload"
	utils2 "github.com/ontio/ontology/core/utils"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/types"
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
	msg := &xshard_state.XShardNotify{
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
	msg := &xshard_state.XShardTxReq{
		IdxInTx:  uint64(reqIdx),
		Payer:    ctx.Tx.Payer,
		Fee:      neovm.MIN_TRANSACTION_GAS,
		Contract: reqParam.ToContract,
		Method:   reqParam.Method,
		Args:     reqParam.Args,
	}
	txState.NextReqID += 1

	if reqIdx < uint32(len(txState.OutReqResp)) {
		if xshard_state.IsXShardMsgEqual(msg, txState.OutReqResp[reqIdx].Req) == false {
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

	txState.PendingReq = msg

	// put Tx-Request
	//todo: clean
	//if err := remoteSendShardMsg(ctx, txHash, reqParam.ToShard, msg); err != nil {
	//	return utils.BYTE_FALSE, fmt.Errorf("remote invoke, notify: %s", err)
	//}

	return utils.BYTE_FALSE, shardsysmsg.ErrYield
}

var ShardA = types.NewShardIDUnchecked(1)
var ShardB = types.NewShardIDUnchecked(2)

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

func RemoteInvokeAdd(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)

	params := shardsysmsg.NotifyReqParam{
		ToShard: ShardB,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	sum, err := RemoteInvoke(native, params)
	return sum, err
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
