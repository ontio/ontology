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
package testsuite

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	utils2 "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/stretchr/testify/assert"
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

func executeTransaction(tx *types.Transaction, cache *storage.CacheDB) (*xshard_state.TxState,
	interface{}, error) {
	config := &smartcontract.Config{
		ShardID: common.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	txHash := tx.Hash()
	txState := xshard_state.CreateTxState(xshard_types.ShardTxID(string(txHash[:])))

	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:       config,
			Store:        nil,
			ShardTxState: txState,
			CacheDB:      cache,
			Gas:          100000000000000,
			PreExec:      true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingOutReq != nil {
				return txState, nil, xshard_state.ErrYield
			}
			return nil, nil, err
		}

		return txState, res, nil
	}

	panic("unimplemented")
}

func ExecuteShardCommandLocal(shardID common.ShardID, command ShardCommand) []byte {
	var result []byte
	switch cmd := command.(type) {
	case *MutliCommand:
		for _, sub := range cmd.SubCmds {
			result = append(result, ExecuteShardCommandLocal(shardID, sub)...)
		}
	case *NotifyCommand:
	case *InvokeCommand:
		result = append(result, ExecuteShardCommandLocal(cmd.Target, cmd.Cmd)...)
	case *GreetCommand:
		result = append(result, fmt.Sprintf("hi from:%d", shardID)...)
	}

	return result
}

func ExecuteShardCommandApi(native *native.NativeService) ([]byte, error) {
	buf, _, _, eof := common.NewZeroCopySource(native.Input).NextVarBytes()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	cmd, err := DecodeShardCommand(common.NewZeroCopySource(buf))
	if err != nil {
		panic(err)
	}

	result, err := ExecuteShardCommand(native, cmd)
	if err != nil {
		return nil, err
	}

	expected := ExecuteShardCommandLocal(native.ShardID, cmd)

	if string(expected) != string(result) {
		panic(fmt.Errorf("execute result mismatch: expected: %s, got: %s", string(expected), string(result)))
	}

	fmt.Println("execute shard command success at shard", native.ShardID)

	return result, nil
}

func ExecuteShardCommand(native *native.NativeService, command ShardCommand) ([]byte, error) {
	cont := native.ContextRef.CurrentContext().ContractAddress
	var result []byte
	switch cmd := command.(type) {
	case *MutliCommand:
		for _, sub := range cmd.SubCmds {
			res, err := ExecuteShardCommand(native, sub)
			if err != nil {
				return nil, err
			}
			result = append(result, res...)
		}
	case *NotifyCommand:
		native.NotifyRemoteShard(cmd.Target, cont, "executeShardCommand", EncodeShardCommandToBytes(cmd.Cmd))
	case *InvokeCommand:
		res, err := native.InvokeRemoteShard(cmd.Target, cont, "executeShardCommand", EncodeShardCommandToBytes(cmd.Cmd))
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	case *GreetCommand:
		result = append(result, fmt.Sprintf("hi from:%d", native.ShardID)...)
	default:
		panic("unkown command")
	}

	return result, nil
}

var ShardA = common.NewShardIDUnchecked(1)
var ShardB = common.NewShardIDUnchecked(2)

func RemoteNotifyPing(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteString(fmt.Sprintf("hello from shard: %d", native.ShardID.ToUint64()))

	cont := native.ContextRef.CurrentContext().ContractAddress
	native.NotifyRemoteShard(ShardB, cont, "handlePing", sink.Bytes())

	return utils.BYTE_TRUE, nil
}

type ReverseStringParam struct {
	Origin []byte
	Shards []common.ShardID
}

func (self *ReverseStringParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(self.Origin)
	sink.WriteUint32(uint32(len(self.Shards)))
	for _, id := range self.Shards {
		sink.WriteShardID(id)
	}
}

func (self *ReverseStringParam) Deserialization(source *common.ZeroCopySource) error {
	origin, _, _, eof := source.NextVarBytes()
	len, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	self.Origin = origin
	for i := uint32(0); i < len; i++ {
		id, err := source.NextShardID()
		if err != nil {
			return err
		}
		self.Shards = append(self.Shards, id)
	}

	return nil
}

func reverse(s []byte) []byte {
	buf := make([]byte, len(s))
	copy(buf, s)
	s = buf
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func ShardReverseBytes(native *native.NativeService) ([]byte, error) {
	buf, _, _, eof := common.NewZeroCopySource(native.Input).NextVarBytes()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	params := &ReverseStringParam{}
	err := params.Deserialization(common.NewZeroCopySource(buf))
	if err != nil {
		panic(err)
		return nil, err
	}

	origin := params.Origin
	var result []byte

	expected := reverse(origin)
	if len(params.Shards) == 0 {
		return expected, nil
	}

	split := len(params.Origin) / len(params.Shards)
	for _, shard := range params.Shards {
		param := origin[:split]
		origin = origin[split:]

		cont := native.ContextRef.CurrentContext().ContractAddress
		method := "shardReverseBytes"
		args := common.SerializeToBytes(&ReverseStringParam{Origin: param})
		res, err := native.InvokeRemoteShard(shard, cont, method, args)
		if err != nil {
			return nil, err
		}
		result = append(res, result...)
	}

	if len(origin) != 0 {
		result = append(reverse(origin), result...)
	}

	if string(expected) != string(result) {
		panic(fmt.Errorf("reverse bytes error: expected:%s, found:%s", string(expected), string(result)))
	}

	return result, nil
}

func RemoteInvokeAddAndInc(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)

	cont := native.ContextRef.CurrentContext().ContractAddress
	sum, err := native.InvokeRemoteShard(ShardB, cont, "handlePing", sink.Bytes())
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

	pushEvent(native, sink.Bytes())

	return sink.Bytes(), err
}

func pushEvent(native *native.NativeService, s interface{}) {
	event := new(event.NotifyEventInfo)
	event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
	event.States = s
	native.Notifications = append(native.Notifications, event)
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

type ShardContext struct {
	shardID         common.ShardID
	ContractAddress common.Address
	overlay         *overlaydb.OverlayDB
	height          uint32
	t               *testing.T
}

func NewShardContext(shardID common.ShardID, contract common.Address, t *testing.T) *ShardContext {
	return &ShardContext{
		shardID:         shardID,
		ContractAddress: contract,
		t:               t,
		overlay:         NewOverlayDB(),
	}
}

func (self *ShardContext) InvokeShardContract(method string, args []interface{}) (common.Uint256, *event.TransactionNotify) {
	t := self.t
	if len(args) == 0 {
		args = []interface{}{""}
	}
	tx := BuildInvokeTx(self.ContractAddress, method, args)
	assert.NotNil(t, tx)
	cache := storage.NewCacheDB(self.overlay)
	xshardDB := storage.NewXShardDB(self.overlay)
	header := &types.Header{Height: self.height, ShardID: self.shardID.ToUint64()}
	self.height += 1
	txHash := tx.Hash()
	txevent := &event.ExecuteNotify{TxHash: txHash, State: event.CONTRACT_STATE_FAIL}
	notify := &event.TransactionNotify{
		ContractEvent: txevent,
	}

	_, err := ledgerstore.HandleInvokeTransaction(nil, self.overlay, cache, xshardDB, tx, header, notify)
	assert.Nil(t, err)
	xshardDB.Commit()

	return txHash, notify
}

func (self *ShardContext) HandleShardCallMsgs(msgs []xshard_types.CommonShardMsg) *event.TransactionNotify {
	t := self.t
	cache := storage.NewCacheDB(self.overlay)
	xshardDB := storage.NewXShardDB(self.overlay)
	header := &types.Header{Height: self.height, ShardID: self.shardID.ToUint64()}
	self.height += 1
	txevent := &event.ExecuteNotify{}
	txevent.State = event.CONTRACT_STATE_FAIL
	notify := &event.TransactionNotify{
		ContractEvent: txevent,
	}
	err := ledgerstore.HandleShardCallTransaction(nil, self.overlay, cache, xshardDB, msgs, header, notify)
	assert.Nil(t, err)
	xshardDB.Commit()

	return notify
}

func (self *ShardContext) GetXShardState(id xshard_types.ShardTxID) (*xshard_state.TxState, error) {
	xshardDB := storage.NewXShardDB(self.overlay)
	return xshardDB.GetXShardState(id)
}

func type2string(ty uint32) string {
	switch ty {
	case xshard_types.EVENT_SHARD_MSG_COMMON:
		return "common"
	case xshard_types.EVENT_SHARD_NOTIFY:
		return "notify"
	case xshard_types.EVENT_SHARD_TXREQ:
		return "request"
	case xshard_types.EVENT_SHARD_TXRSP:
		return "response"
	case xshard_types.EVENT_SHARD_PREPARE:
		return "prepare"
	case xshard_types.EVENT_SHARD_PREPARED:
		return "prepared"
	case xshard_types.EVENT_SHARD_COMMIT:
		return "commit"
	case xshard_types.EVENT_SHARD_ABORT:
		return "abort"
	}

	return "unkown type"
}

func GetShardMsgInfo(msg xshard_types.CommonShardMsg) string {
	ty := type2string(msg.Type())
	return fmt.Sprintf("shardmsg: %s, shard%d->shard%d : shardid: %x", ty, msg.GetSourceShardID(),
		msg.GetTargetShardID(), msg.GetShardTxID())
}

func RunShardTxToComplete(shards map[common.ShardID]*ShardContext, shard common.ShardID, method string, args []byte) int {
	_, notify := shards[shard].InvokeShardContract(method, []interface{}{args})
	shardMsgs := notify.ShardMsg
	totalShardMsg := 0
	for len(shardMsgs) != 0 {
		msg := shardMsgs[0]
		shardMsgs = shardMsgs[1:]
		msgs := []xshard_types.CommonShardMsg{msg}
		target := msg.GetTargetShardID()
		fmt.Println(GetShardMsgInfo(msg))
		notify := shards[target].HandleShardCallMsgs(msgs)
		shardMsgs = append(shardMsgs, notify.ShardMsg...)
		totalShardMsg += 1
	}

	return totalShardMsg
}
