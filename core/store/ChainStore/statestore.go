package ChainStore

import (
	"bytes"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/core/states"
	. "github.com/Ontology/core/store"
	"github.com/Ontology/core/store/statestore"
	"strings"
)

type StateStore struct {
	db          *ChainStore
	memoryStore IMemoryStore
	trieStore   statestore.ITrieStore
	trie        statestore.ITrie
}

func NewStateStore(memoryStore IMemoryStore, db *ChainStore, root common.Uint256) *StateStore {
	trieStore := statestore.NewTrieStore(db.st)
	tr, err := trieStore.OpenTrie(root)
	if err != nil {
		panic("[NewStateStore] opentrie error:" + err.Error())
	}
	return &StateStore{
		db:          db,
		memoryStore: memoryStore,
		trieStore:   trieStore,
		trie:        tr,
	}
}

func (self *StateStore) Find(prefix DataEntryPrefix, key []byte) ([]*StateItem, error) {
	var states []*StateItem
	iter := self.db.st.NewIterator(append([]byte{byte(prefix)}, key...))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		state, err := getStateObject(prefix, value)
		if err != nil {
			return nil, err
		}
		states = append(states, &StateItem{Key: string(key[1:]), Value: state})
	}
	return states, nil
}

func (self *StateStore) TryAdd(prefix DataEntryPrefix, key []byte, value IStateValue, trie bool) {
	self.setStateObject(byte(prefix), key, value, Changed, trie)
}

func (self *StateStore) TryGetOrAdd(prefix DataEntryPrefix, key []byte, value IStateValue, trie bool) error {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == Deleted {
			self.setStateObject(byte(prefix), key, value, Changed, trie)
			return nil
		}
		return nil
	}
	item, err := self.db.st.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && !strings.EqualFold(err.Error(), ErrDBNotFound) {
		return err
	}
	if item != nil {
		return nil
	}
	self.setStateObject(byte(prefix), key, value, Changed, trie)
	return nil
}

func (self *StateStore) TryGet(prefix DataEntryPrefix, key []byte) (*StateItem, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == Deleted {
			return nil, nil
		}
		return state, nil
	}
	enc, err := self.db.st.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && !strings.EqualFold(err.Error(), ErrDBNotFound) {
		return nil, err
	}

	if enc == nil {
		return nil, nil
	}
	stateVal, err := getStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	self.setStateObject(byte(prefix), key, stateVal, None, false)
	return &StateItem{Key: string(append([]byte{byte(prefix)}, key...)), Value: stateVal, State: None}, nil
}

func (self *StateStore) TryGetAndChange(prefix DataEntryPrefix, key []byte, trie bool) (IStateValue, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == Deleted {
			return nil, nil
		} else if state.State == None {
			state.State = Changed
		}
		return state.Value, nil
	}
	k := append([]byte{byte(prefix)}, key...)
	enc, err := self.db.st.Get(k)
	if err != nil && !strings.EqualFold(err.Error(), ErrDBNotFound) {
		return nil, err
	}

	if enc == nil {
		return nil, nil
	}

	val, err := getStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	self.setStateObject(byte(prefix), key, val, Changed, trie)
	return val, nil
}

func (self *StateStore) TryDelete(prefix DataEntryPrefix, key []byte) {
	self.memoryStore.Delete(byte(prefix), key)
}

func (self *StateStore) CommitTo() error {
	for k, v := range self.memoryStore.GetChangeSet() {
		if v.State == Deleted {
			if v.Trie {
				if err := self.trie.TryDelete([]byte(k)); err != nil {
					return err
				}
			}
			if err := self.db.st.BatchDelete([]byte(k)); err != nil {
				return err
			}
		} else {
			data := new(bytes.Buffer)
			err := v.Value.Serialize(data)
			if err != nil {
				log.Errorf("[CommitTo] error: key %v, value:%v", k, v.Value)
				return err
			}
			if v.Trie {
				value := common.ToHash256(data.Bytes())
				if err := self.trie.TryUpdate([]byte(k), value.ToArray()); err != nil {
					return err
				}
			}
			if err = self.db.st.BatchPut([]byte(k), data.Bytes()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *StateStore) setStateObject(prefix byte, key []byte, value IStateValue, state ItemState, trie bool) {
	self.memoryStore.Put(prefix, key, value, state, trie)
}

func getStateObject(prefix DataEntryPrefix, enc []byte) (IStateValue, error) {
	reader := bytes.NewBuffer(enc)
	switch prefix {
	case ST_Account:
		account := new(AccountState)
		if err := account.Deserialize(reader); err != nil {
			return nil, err
		}
		return account, nil
	case ST_Coin:
		unspentcoin := new(UnspentCoinState)
		if err := unspentcoin.Deserialize(reader); err != nil {
			return nil, err
		}
		return unspentcoin, nil
	case ST_SpentCoin:
		spentcoin := new(SpentCoinState)
		if err := spentcoin.Deserialize(reader); err != nil {
			return nil, err
		}
		return spentcoin, nil
	case ST_BookKeeper:
		bookKeeper := new(BookKeeperState)
		if err := bookKeeper.Deserialize(reader); err != nil {
			return nil, err
		}
		return bookKeeper, nil
	case ST_Asset:
		asset := new(AssetState)
		if err := asset.Deserialize(reader); err != nil {
			return nil, err
		}
		return asset, nil
	case ST_Contract:
		contract := new(ContractState)
		if err := contract.Deserialize(reader); err != nil {
			return nil, err
		}
		return contract, nil
	case ST_Storage:
		storage := new(StorageItem)
		if err := storage.Deserialize(reader); err != nil {
			return nil, err
		}
		return storage, nil
	case ST_Program_Coin:
		programCoin := new(ProgramUnspentCoin)
		if err := programCoin.Deserialize(reader); err != nil {
			return nil, err
		}
		return programCoin, nil
	default:
		panic("[getStateObject] invalid state type!")
	}
}

func newStateObject(prefix DataEntryPrefix) IStateValue {
	switch prefix {
	case ST_Account:
		return new(AccountState)
	case ST_Coin:
		return new(UnspentCoinState)
	case ST_SpentCoin:
		return new(SpentCoinState)
	case ST_BookKeeper:
		return new(BookKeeperState)
	case ST_Asset:
		return new(AssetState)
	case ST_Contract:
		return new(ContractState)
	case ST_Storage:
		return new(StorageItem)
	default:
		panic("[newStateObject] invalid state type!")
	}
}
