package ledger

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	"github.com/Ontology/core/store/ledgerstore"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/core/payload"
)

var DefLedger *Ledger

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	ldgStore store.ILedgerStore
}

func NewLedger() (*Ledger, error) {
	ldgStore, err := ledgerstore.NewLedgerStore()
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	return &Ledger{
		ldgStore: ldgStore,
	}, nil
}

func (this *Ledger) GetStore() store.ILedgerStore {
	return this.ldgStore
}

func (this *Ledger) Init(defaultBookKeeper []*crypto.PubKey) error {
	genesisBlock, err := genesis.GenesisBlockInit(defaultBookKeeper)
	if err != nil {
		return fmt.Errorf("genesisBlock error %s", err)
	}
	err = this.ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock, defaultBookKeeper)
	if err != nil {
		return fmt.Errorf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	return nil
}

func (this *Ledger) AddHeaders(headers []*types.Header) error {
	return this.ldgStore.AddHeaders(headers)
}

func (this *Ledger) AddBlock(block *types.Block) error {
	return this.ldgStore.AddBlock(block)
}

func (this *Ledger) GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256 {
	return this.ldgStore.GetBlockRootWithNewTxRoot(txRoot)
}

func (this *Ledger) GetBlockByHeight(height uint32) (*types.Block, error) {
	return this.ldgStore.GetBlockByHeight(height)
}

func (this *Ledger) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return this.ldgStore.GetBlockByHash(blockHash)
}

func (this *Ledger) GetHeaderByHeight(height uint32) (*types.Header, error) {
	return this.ldgStore.GetHeaderByHeight(height)
}

func (this *Ledger) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	return this.ldgStore.GetHeaderByHash(blockHash)
}

func (this *Ledger) GetBlockHash(height uint32) common.Uint256 {
	return this.ldgStore.GetBlockHash(height)
}

func (this *Ledger) GetTransaction(txHash common.Uint256) (*types.Transaction, error) {
	tx, _, err := this.ldgStore.GetTransaction(txHash)
	return tx, err
}

func (this *Ledger) GetTransactionWithHeight(txHash common.Uint256) (*types.Transaction, uint32, error) {
	return this.ldgStore.GetTransaction(txHash)
}

func (this *Ledger) GetCurrentBlockHeight() uint32 {
	return this.ldgStore.GetCurrentBlockHeight()
}

func (this *Ledger) GetCurrentBlockHash() common.Uint256 {
	return this.ldgStore.GetCurrentBlockHash()
}

func (this *Ledger) GetCurrentHeaderHeight() uint32 {
	return this.ldgStore.GetCurrentHeaderHeight()
}

func (this *Ledger) GetCurrentHeaderHash() common.Uint256 {
	return this.ldgStore.GetCurrentHeaderHash()
}

func (this *Ledger) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return this.ldgStore.IsContainTransaction(txHash)
}

func (this *Ledger) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return this.ldgStore.IsContainBlock(blockHash)
}

func (this *Ledger) IsDoubleSpend(tx *types.Transaction) (bool, error) {
	//txInputs := tx.UTXOInputs
	//if len(txInputs) == 0 {
	//	return false, nil
	//}
	//groups := this.groupInputIndex(txInputs)
	//for refTx, intputs := range groups {
	//	coinState, err := this.ldgStore.GetUnspentCoinState(&refTx)
	//	if err != nil {
	//		return false, fmt.Errorf("GetUnspentCoinState tx %x error %s", refTx, err)
	//	}
	//	if coinState == nil {
	//		return true, nil
	//	}
	//	for _, index := range intputs {
	//		if int(index) >= len(coinState.Item) || coinState.Item[index] == states.Spent {
	//			return true, nil
	//		}
	//	}
	//}
	return false, nil
}

func (this *Ledger) GetCurrentStateRoot() (common.Uint256, error) {
	return common.Uint256{}, nil
}

func (this *Ledger) GetBookKeeperState() (*states.BookKeeperState, error) {
	return this.ldgStore.GetBookKeeperState()
}

//
//func (this *Ledger) GetBookKeeperAddress() (*common.Address, error) {
//	bookState, err := this.GetBookKeeperState()
//	if err != nil {
//		return nil, fmt.Errorf("GetBookKeeperState error %s", err)
//	}
//	address, err := this.MakeBookKeeperAddress(bookState.CurrBookKeeper)
//	if err != nil {
//		return nil, fmt.Errorf("makeBookKeeperAddress error %s", err)
//	}
//	return address, err
//}

//func (this *Ledger) GetUnclaimed(txHash common.Uint256) (map[uint16]*utxo.SpentCoin, error) {
//tx, err := this.GetTransaction(txHash)
//if err != nil {
//	return nil, fmt.Errorf("GetTransaction error %s", err)
//}
//if tx == nil {
//	return nil, fmt.Errorf("cannot get tx by %x", txHash)
//}
//spentState, err := this.ldgStore.GetSpentCoinState(txHash)
//if err != nil {
//	return nil, fmt.Errorf("GetSpentCoinState error %s", err)
//}
//if spentState == nil {
//	return nil, fmt.Errorf("cannot get spent coin state by %x", tx)
//}
//claimable := make(map[uint16]*utxo.SpentCoin, len(spentState.Items))
//for _, spent := range spentState.Items{
//	index := spent.PrevIndex
//	claimable[index] = &utxo.SpentCoin{
//		Output:tx.Outputs[index],
//		StartHeight:spentState.TransactionHeight,
//		EndHeight:spent.EndHeight,
//	}
//}
//return claimable, nil
//}

func (this *Ledger) GetStorageItem(codeHash *common.Address, key []byte) ([]byte, error) {
	storageKey := &states.StorageKey{
		CodeHash: *codeHash,
		Key:      key,
	}
	storageItem, err := this.ldgStore.GetStorageItem(storageKey)
	if err != nil {
		return nil, fmt.Errorf("GetStorageItem error %s", err)
	}
	if storageItem == nil {
		return nil, fmt.Errorf("cannot get storage item by codehash %x Key %x", *codeHash, key)
	}
	return storageItem.Value, nil
}

func (this *Ledger) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	return this.ldgStore.GetContractState(contractHash)
}

func (this *Ledger) PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
	return this.ldgStore.PreExecuteContract(tx)
}

//
//func (this *Ledger) GetAllAssetState() (map[common.Uint256]*states.AssetState, error) {
//	return this.ldgStore.GetAllAssetState()
//}
//
//func (this *Ledger) MakeBookKeeperAddress(bookKeepers []*crypto.PubKey) (*common.Address, error) {
//	bookSize := len(bookKeepers)
//	if bookSize == 0 {
//		return nil, fmt.Errorf("bookKeeper is empty")
//	}
//	var script []byte
//	var err error
//	if bookSize == 1 {
//		script, err = contract.CreateSignatureRedeemScript(bookKeepers[0])
//		if err != nil {
//			return nil, fmt.Errorf("CreateSignatureRedeemScript error %s", err)
//		}
//	} else {
//		script, err = contract.CreateMultiSigRedeemScript(bookSize-(bookSize-1)/3, bookKeepers)
//		if err != nil {
//			return nil, fmt.Errorf("CreateMultiSigRedeemScript error %s", err)
//		}
//	}
//	codeHash, err := common.ToCodeHash(script)
//	if err != nil {
//		return nil, fmt.Errorf("ToCodeHash error %s", err)
//	}
//	return &codeHash, nil
//}
