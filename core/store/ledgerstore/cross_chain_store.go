package ledgerstore

import (
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
	sink := common.NewZeroCopySink(nil)
	key := this.genCrossChainMsgKey(sink, crossChainMsg.Height)
	sink.Reset()
	if err := crossChainMsg.Serialization(sink); err != nil {
		return err
	}
	this.store.Put(key, sink.Bytes())
	return nil
}

func (this *CrossChainStore) genCrossChainMsgKey(sink *common.ZeroCopySink, height uint32) []byte {
	sink.WriteByte(byte(scom.SYS_CROSS_CHAIN_MSG))
	sink.WriteUint32(height)
	return sink.Bytes()
}
