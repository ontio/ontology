package ChainStore

import (
	"bytes"
	"encoding/binary"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/ledger"
	. "github.com/Ontology/core/states"
	. "github.com/Ontology/core/store"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/core/transaction/utxo"
	"math/big"
	"github.com/Ontology/crypto"
)

func repeat(length int) []CoinState {
	items := make([]CoinState, length)
	for i := 0; i < length; i++ {
		items[i] = Confirmed
	}
	return items
}

func handleOutputs(txid Uint256, outputs []*utxo.TxOutput, stateStore *StateStore) error {
	for i, o := range outputs {
		state, err := stateStore.TryGetAndChange(ST_Account, o.ProgramHash.ToArray(), true)
		if err != nil {
			log.Errorf("[handleOutputs] TryGetAndChange ST_Account error: %v", err)
			return err
		}
		ph := o.ProgramHash
		as := o.AssetID
		if state == nil {
			balances := make(map[Uint256]Fixed64)
			balances[as] = o.Value
			stateStore.TryAdd(ST_Account, ph.ToArray(), &AccountState{ProgramHash: ph, IsFrozen: false, Balances: balances}, true)
		} else {
			account := state.(*AccountState)
			account.Balances[as] += o.Value
		}
		state, err = stateStore.TryGetAndChange(ST_Program_Coin, append(ph.ToArray(), as.ToArray()...), false)
		if err != nil {
			log.Errorf("[handleOutputs] TryGetAndChange ST_Program_Coin error: %v", err)
			return err
		}
		unspent := &utxo.UTXOUnspent{Txid: txid, Index: uint32(i), Value: o.Value}
		if state == nil {
			stateStore.TryAdd(ST_Program_Coin, append(ph.ToArray(), as.ToArray()...), &ProgramUnspentCoin{Unspents: []*utxo.UTXOUnspent{unspent}}, false)
		} else {
			programCoin := state.(*ProgramUnspentCoin)
			programCoin.Unspents = append(programCoin.Unspents, unspent)
		}
	}
	return nil
}

func handleInputs(inputs []*utxo.UTXOTxInput, stateStore *StateStore, currentBlockHeight uint32, bd *ChainStore) error {
	for _, i := range inputs {
		tx_prev := new(tx.Transaction)
		refer_tx := i.ReferTxID.ToArray()
		height, err := bd.getTx(tx_prev, i.ReferTxID)
		if err != nil {
			log.Errorf("[persist] getTx error: %v", err)
			return err
		}

		state, err := stateStore.TryGetAndChange(ST_Coin, refer_tx, false)
		if err != nil {
			log.Errorf("[persist] TryGet ST_Coin error:", err)
			return err
		}
		unspentcoins := state.(*UnspentCoinState)
		unspentcoins.Item[i.ReferTxOutputIndex] = Spent

		state, err = stateStore.TryGetAndChange(ST_SpentCoin, refer_tx, false)
		if err != nil {
			log.Errorf("[persist] TryGet ST_SpentCoin error:", err)
			return err
		}
		if state == nil {
			items := make([]*Item, 0)
			items = append(items, &Item{PrevIndex: i.ReferTxOutputIndex, EndHeight: currentBlockHeight})
			stateStore.TryAdd(ST_SpentCoin, refer_tx, &SpentCoinState{TransactionHash: i.ReferTxID, TransactionHeight: height, Items: items}, false)
		} else {
			spentcoin := state.(*SpentCoinState)
			spentcoin.Items = append(spentcoin.Items, &Item{PrevIndex: i.ReferTxOutputIndex, EndHeight: currentBlockHeight})
		}

		prev_output := tx_prev.Outputs[i.ReferTxOutputIndex]
		ph := prev_output.ProgramHash.ToArray()
		state, err = stateStore.TryGetAndChange(ST_Account, ph, true)
		if err != nil {
			log.Errorf("[persist] TryGet ST_Account error: %v", err)
			return err
		}
		account := state.(*AccountState)
		account.Balances[prev_output.AssetID] -= prev_output.Value

		state, err = stateStore.TryGetAndChange(ST_Program_Coin, append(ph, prev_output.AssetID.ToArray()...), false)
		if err != nil {
			log.Errorf("[handleOutputs] TryGetAndChange ST_Program_Coin error: %v", err)
			return err
		}
		programCoin := state.(*ProgramUnspentCoin)
		programCoin.Unspents = append(programCoin.Unspents[:i.ReferTxOutputIndex], programCoin.Unspents[i.ReferTxOutputIndex:]...)
	}
	return nil
}

func handleBookKeeper(stateStore *StateStore, bookKeeper *BookKeeperState) {
	flag := false
	if len(bookKeeper.CurrBookKeeper) != len(bookKeeper.NextBookKeeper) {
		flag = true
	} else {
		for i := range bookKeeper.CurrBookKeeper {
			if bookKeeper.CurrBookKeeper[i].X.Cmp(bookKeeper.NextBookKeeper[i].X) != 0 ||
				bookKeeper.CurrBookKeeper[i].Y.Cmp(bookKeeper.NextBookKeeper[i].Y) != 0 {
				flag = true
				break
			}
		}
	}
	if flag {
		bookKeeper.CurrBookKeeper = make([]*crypto.PubKey, len(bookKeeper.NextBookKeeper))
		for i := 0; i < len(bookKeeper.NextBookKeeper); i++ {
			bookKeeper.CurrBookKeeper[i] = new(crypto.PubKey)
			bookKeeper.CurrBookKeeper[i].X = new(big.Int).Set(bookKeeper.NextBookKeeper[i].X)
			bookKeeper.CurrBookKeeper[i].Y = new(big.Int).Set(bookKeeper.NextBookKeeper[i].Y)
		}
		stateStore.memoryStore.Change(byte(ST_BookKeeper), BookerKeeper, false)
	}
}

func addHeader(bd *ChainStore, b *ledger.Block, curr_block_sysfee uint64) error {
	key := bytes.NewBuffer(append([]byte{byte(DATA_Header)}))
	bh := b.Hash()
	if _, err := bh.Serialize(key); err != nil {
		return err
	}
	value := new(bytes.Buffer)

	sysfee := Fixed64(0)
	if err := sysfee.Serialize(value); err != nil {
		return err
	}
	b.Trim(value)
	bd.st.BatchPut(key.Bytes(), value.Bytes())
	return nil
}

func addSysCurrentBlock(bd *ChainStore, b *ledger.Block) error {
	key := bytes.NewBuffer(append([]byte{byte(SYS_CurrentBlock)}))
	value := new(bytes.Buffer)
	bh := b.Hash()
	if _, err := bh.Serialize(value); err != nil {
		return err
	}
	if err := serialization.WriteUint32(value, b.Blockdata.Height); err != nil {
		return err
	}
	bd.st.BatchPut(key.Bytes(), value.Bytes())
	return nil
}

func addDataBlock(bd *ChainStore, b *ledger.Block) error {
	key := bytes.NewBuffer(append([]byte{byte(DATA_Block)}))
	if err := serialization.WriteUint32(key, b.Blockdata.Height); err != nil {
		return err
	}
	value := new(bytes.Buffer)
	bh := b.Hash()
	if _, err := bh.Serialize(value); err != nil {
		return err
	}
	bd.st.BatchPut(key.Bytes(), value.Bytes())
	return nil
}

func addCurrentStateRoot(bd *ChainStore, stateRoot Uint256) error {
	return bd.st.BatchPut(append([]byte{byte(Sys_CurrentStateRoot)}, CurrentStateRoot...), stateRoot.ToArray())
}

func addMerkleRoot(bd *ChainStore, b *ledger.Block) {
	// update merkle tree
	bd.merkleTree.AppendHash(b.Blockdata.TransactionsRoot)
	bd.merkleHashStore.Flush()

	tree_size := bd.merkleTree.TreeSize()
	hashes := bd.merkleTree.Hashes()
	length := 4 + len(hashes) * UINT256SIZE
	buf := make([]byte, 4, length)
	binary.BigEndian.PutUint32(buf[0:], tree_size)
	for _, h := range hashes {
		buf = append(buf, h[:]...)
	}
	bd.st.BatchPut([]byte{byte(SYS_BlockMerkleTree)}, buf)
}

func groupInputs(inputs []*utxo.UTXOTxInput) map[Uint256][]*utxo.UTXOTxInput {
	group := make(map[Uint256][]*utxo.UTXOTxInput)
	for _, v := range inputs {
		group[v.ReferTxID] = append(group[v.ReferTxID], v)
	}
	return group
}

func remove(items []*Item, index int) []*Item {
	return append(items[:index], items[index + 1:]...)
}
