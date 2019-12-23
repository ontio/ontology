package ledgerstore

import (
	"encoding/binary"
	"fmt"
	"github.com/ontio/ontology/common"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
	"os"
)

const (
	DBDirCrossChain = "crosschain"
)

//Block store save the data of block & transaction
type CrossChainStore struct {
	dbDir                 string                     //The path of store file
	store                 *leveldbstore.LevelDBStore //block store handler
	crossChainCheckHeight uint32                     //cross chain start height
}

//NewCrossChainStore return cross chain store instance
func NewCrossChainStore(dataDir string, crossChainHeight uint32) (*CrossChainStore, error) {
	dbDir := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirCrossChain)
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, fmt.Errorf("NewCrossShardStore error %s", err)
	}
	return &CrossChainStore{
		dbDir:                 dbDir,
		store:                 store,
		crossChainCheckHeight: crossChainHeight,
	}, nil
}

func (this *CrossChainStore) SaveMsgToCrossChainStore(crossChainMsg *types.CrossChainMsg) error {
	if crossChainMsg == nil || crossChainMsg.Height < this.crossChainCheckHeight {
		return nil
	}
	key := this.genCrossChainMsgKey(crossChainMsg.Height)
	sink := common.NewZeroCopySink(nil)
	crossChainMsg.Serialization(sink)
	this.store.Put(key, sink.Bytes())
	return nil
}

func (this *CrossChainStore) GetCrossChainMsg(height uint32) (*types.CrossChainMsg, error) {
	key := this.genCrossChainMsgKey(height)
	value, err := this.store.Get(key)
	if err != nil && err != scom.ErrNotFound {
		return nil, err
	}
	if err == scom.ErrNotFound {
		return nil, nil
	}
	source := common.NewZeroCopySource(value)
	msg := new(types.CrossChainMsg)
	if err := msg.Deserialization(source); err != nil {
		return nil, err
	}
	return msg, nil
}

func (this *CrossChainStore) genCrossChainMsgKey(height uint32) []byte {
	temp := make([]byte, 5)
	temp[0] = byte(scom.SYS_CROSS_CHAIN_MSG)
	binary.LittleEndian.PutUint32(temp[1:], height)
	return temp
}
