package common

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/consensus"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
	. "github.com/Ontology/net/protocol"
	"strconv"
)

var CNoder Noder
var ConsensusSrv ConsensusService

//multiplexer that keeps track of every function to be called on specific rpc call

type TxAttributeInfo struct {
	Usage types.TransactionAttributeUsage
	Data  string
}

type UTXOTxInputInfo struct {
	ReferTxID          string
	ReferTxOutputIndex uint16
}

type BalanceTxInputInfo struct {
	AssetID     string
	Value       string
	ProgramHash string
}

type TxoutputInfo struct {
	AssetID     string
	Value       string
	ProgramHash string
}

type TxoutputMap struct {
	Key   Uint256
	Txout []TxoutputInfo
}

type AmountMap struct {
	Key   Uint256
	Value Fixed64
}

type ProgramInfo struct {
	Code      string
	Parameter string
}

type Transactions struct {
	TxType         types.TransactionType
	PayloadVersion byte
	Payload        PayloadInfo
	Attributes     []TxAttributeInfo
	UTXOInputs     []UTXOTxInputInfo
	BalanceInputs  []BalanceTxInputInfo
	Outputs        []TxoutputInfo
	Programs       []ProgramInfo
	NetworkFee     string
	SystemFee      string

	AssetOutputs      []TxoutputMap
	AssetInputAmount  []AmountMap
	AssetOutputAmount []AmountMap

	Hash string
}

type BlockHead struct {
	Version          uint32
	PrevBlockHash    string
	TransactionsRoot string
	BlockRoot        string
	StateRoot        string
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	NextBookKeeper   string
	Program          ProgramInfo

	Hash string
}

type BlockInfo struct {
	Hash         string
	BlockData    *BlockHead
	Transactions []*Transactions
}

type TxInfo struct {
	Hash string
	Hex  string
	Tx   *Transactions
}

type TxoutInfo struct {
	High  uint32
	Low   uint32
	Txout utxo.TxOutput
}

type NodeInfo struct {
	State    uint   // node status
	Port     uint16 // The nodes's port
	ID       uint64 // The nodes's id
	Time     int64
	Version  uint32 // The network protocol the node used
	Services uint64 // The services the node supplied
	Relay    bool   // The relay capability of the node (merge into capbility flag)
	Height   uint64 // The node latest block height
	TxnCnt   uint64 // The transactions be transmit by this node
	RxTxnCnt uint64 // The transaction received by this node
}

type ConsensusInfo struct {
	// TODO
}

func VerifyAndSendTx(txn *types.Transaction) ErrCode {
	// if transaction is verified unsucessfully then will not put it into transaction pool
	if errCode := CNoder.AppendTxnPool(txn); errCode != ErrNoError {
		log.Warn("Can NOT add the transaction to TxnPool")
		log.Info("[httpjsonrpc] VerifyTransaction failed when AppendTxnPool.")
		return errCode
	}
	if err := CNoder.Xmit(txn); err != nil {
		log.Error("Xmit Tx Error:Xmit transaction failed.", err)
		return ErrXmitFail
	}
	return ErrNoError
}
func TransArryByteToHexString(ptx *types.Transaction) *Transactions {
	panic("Transaction structure has changed need reimplement ")

	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.Payload = TransPayloadToHex(ptx.Payload)

	n := 0
	trans.Attributes = make([]TxAttributeInfo, len(ptx.Attributes))
	for _, v := range ptx.Attributes {
		trans.Attributes[n].Usage = v.Usage
		trans.Attributes[n].Data = ToHexString(v.Data)
		n++
	}

	networkfee := ptx.GetNetworkFee()
	trans.NetworkFee = strconv.FormatInt(int64(networkfee), 10)

	mhash := ptx.Hash()
	trans.Hash = ToHexString(mhash.ToArray())

	return trans
}
