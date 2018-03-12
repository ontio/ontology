package rpc

import (
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/common"
	. "github.com/Ontology/http/base/actor"
	"os"
	"path/filepath"
)

const (
	RANDBYTELEN = 4
)

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func GetNeighbor(params []interface{}) map[string]interface{} {
	addr, _ := GetNeighborAddrs()
	return DnaRpc(addr)
}

func GetNodeState(params []interface{}) map[string]interface{} {
	state,err := GetConnectionState()
	if err != nil {
		return DnaRpcFailed
	}
	t,err := GetNodeTime()
	if err != nil {
		return DnaRpcFailed
	}
	port,err := GetNodePort()
	if err != nil {
		return DnaRpcFailed
	}
	id,err := GetID()
	if err != nil {
		return DnaRpcFailed
	}
	ver,err := GetVersion()
	if err != nil {
		return DnaRpcFailed
	}
	tpe,err := GetNodeType()
	if err != nil {
		return DnaRpcFailed
	}
	relay,err := GetRelayState()
	if err != nil {
		return DnaRpcFailed
	}
	height,err := BlockHeight()
	if err != nil {
		return DnaRpcFailed
	}
	txnCnt,err := GetTxnCnt()
	if err != nil {
		return DnaRpcFailed
	}
	n := NodeInfo{
		NodeState:    uint(state),
		NodeTime:     t,
		NodePort:     port,
		ID:       id,
		NodeVersion:  ver,
		NodeType: tpe,
		Relay:    relay,
		Height:   height,
		TxnCnt:   txnCnt,
	}
	return DnaRpc(n)
}

func StartConsensus(params []interface{}) map[string]interface{} {
	if err := ConsensusSrvStart(); err != nil {
		return DnaRpcFailed
	}
	return DnaRpcSuccess
}

func StopConsensus(params []interface{}) map[string]interface{} {
	if err := ConsensusSrvHalt(); err != nil {
		return DnaRpcFailed
	}
	return DnaRpcSuccess
}

func SendSampleTransaction(params []interface{}) map[string]interface{} {
	panic("need reimplementation")
	return nil

	/*
		if len(params) < 1 {
			return DnaRpcNil
		}
		var txType string
		switch params[0].(type) {
		case string:
			txType = params[0].(string)
		default:
			return DnaRpcInvalidParameter
		}

		issuer, err := account.NewAccount()
		if err != nil {
			return DnaRpc("Failed to create account")
		}
		admin := issuer

		rbuf := make([]byte, RANDBYTELEN)
		rand.Read(rbuf)
		switch string(txType) {
		case "perf":
			num := 1
			if len(params) == 2 {
				switch params[1].(type) {
				case float64:
					num = int(params[1].(float64))
				}
			}
			for i := 0; i < num; i++ {
				regTx := NewRegTx(ToHexString(rbuf), i, admin, issuer)
				SignTx(admin, regTx)
				VerifyAndSendTx(regTx)
			}
			return DnaRpc(fmt.Sprintf("%d transaction(s) was sent", num))
		default:
			return DnaRpc("Invalid transacion type")
		}
	*/
}

func SetDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return DnaRpcInvalidParameter
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log.SetDebugLevel(int(level)); err != nil {
			return DnaRpcInvalidParameter
		}
	default:
		return DnaRpcInvalidParameter
	}
	return DnaRpcSuccess
}
