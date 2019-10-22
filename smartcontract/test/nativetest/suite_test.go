package test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	commons "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	common2 "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	utils3 "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	sstate "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
	types2 "github.com/ontio/ontology/vm/neovm/types"
	//"github.com/stretchr/testify/assert"
)

func TestNativeContract(t *testing.T) {
	InstallNativeContract(utils3.OntContractAddress, map[string]native.Handler{
		"OntInitTest": OntInitTest,
		"transfer":    ont.OntTransfer,
	})
	BuildNativeTxsAndTest(t)
}

func BuildNativeTxsAndTest(t *testing.T) {
	var sts []*ont.State
	from, _ := common.AddressFromBase58(string("Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT"))
	to, _ := common.AddressFromBase58(string("Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV"))
	sts = append(sts, &ont.State{
		From:  from,
		To:    to,
		Value: 30,
	})

	cache := NewNativeTxTest(nil, utils3.OntContractAddress, from, "OntInitTest", []interface{}{""})
	NewNativeTxTest(cache, utils3.OntContractAddress, from, ont.TRANSFER_NAME, []interface{}{sts})
}

func OntInitTest(native *native.NativeService) ([]byte, error) {
	from, _ := common.AddressFromBase58(string("Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT"))
	val := uint64(10000)
	balanceKey := ont.GenBalanceKey(utils3.OntContractAddress, from)
	item := utils3.GenUInt64StorageItem(val)
	native.CacheDB.Put(balanceKey, item.ToArray())
	ont.AddNotifications(native, utils3.OntContractAddress, &ont.State{To: from, Value: val})

	return utils3.BYTE_TRUE, nil
}

func NewNativeTxTest(cache *storage.CacheDB, contractAddress common.Address, sigaddr common.Address, method string, params []interface{}) *storage.CacheDB {
	mutable, err := common2.NewNativeInvokeTransaction(0, 2000000000, contractAddress, 0, method, params)
	if err != nil {
		panic(err)
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic(err)
	}
	tx.SignedAddr = append(tx.SignedAddr, sigaddr)

	if cache == nil {
		overlay := NewOverlayDB()
		cache = storage.NewCacheDB(overlay)
	}

	res, err := executeTransaction(tx, cache)
	if err != nil {
		panic(err)
	}

	if res.State == 0 {
		panic("execute error")
	}

	js, _ := json.Marshal(res.Notify)
	log.Infof("Notify info : %s", string(js))
	return cache
}

func executeTransaction(tx *types.Transaction, cache *storage.CacheDB) (*sstate.PreExecResult, error) {
	stf := &sstate.PreExecResult{State: event.CONTRACT_STATE_FAIL, Result: nil}
	sconfig := &smartcontract.Config{
		Time:   uint32(time.Now().Unix()),
		Height: 7000000,
		Tx:     tx,
	}

	gasTable := make(map[string]uint64)
	neovm.GAS_TABLE.Range(func(k, value interface{}) bool {
		key := k.(string)
		val := value.(uint64)
		gasTable[key] = val

		return true
	})
	if tx.TxType == types.InvokeNeo {
		invoke := tx.Payload.(*payload.InvokeCode)
		sc := smartcontract.SmartContract{
			Config:   sconfig,
			CacheDB:  cache,
			GasTable: gasTable,
			Gas:      math.MaxUint64,
			PreExec:  true,
		}
		engine, _ := sc.NewExecuteEngine(invoke.Code, tx.TxType)
		result, err := engine.Invoke()
		if err != nil {
			return stf, err
		}
		var cv interface{}
		if tx.TxType == types.InvokeNeo { //neovm
			if result != nil {
				val := result.(*types2.VmValue)
				cv, err = val.ConvertNeoVmValueHexString()
				if err != nil {
					return stf, err
				}
			}
		} else { //wasmvm
			cv = common.ToHexString(result.([]byte))
		}
		return &sstate.PreExecResult{State: event.CONTRACT_STATE_SUCCESS, Result: cv, Notify: sc.Notifications}, nil
	}

	return stf, errors.NewErr("wrong tx type")
}

func InstallNativeContract(addr common.Address, actions map[string]native.Handler) {
	contract := func(native *native.NativeService) {
		for name, fun := range actions {
			native.Register(name, fun)
		}
	}
	native.Contracts[addr] = contract
}

type MockDB struct {
	commons.PersistStore
	db map[string]string
}

func (self *MockDB) Get(key []byte) ([]byte, error) {
	val, ok := self.db[string(key)]
	if ok == false {
		return nil, commons.ErrNotFound
	}
	return []byte(val), nil
}

func (self *MockDB) BatchPut(key []byte, value []byte) {
	self.db[string(key)] = string(value)
}

func (self *MockDB) BatchDelete(key []byte) {
	delete(self.db, string(key))
}

func NewOverlayDB() *overlaydb.OverlayDB {
	return overlaydb.NewOverlayDB(&MockDB{nil, make(map[string]string)})
}
