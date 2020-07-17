package ledgerstore

import (
	"github.com/ontio/ontology/common"
	"github.com/tendermint/iavl"
)

type StoreProof iavl.RangeProof

func (this *StoreProof) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.LeftPath)))
	for _, item := range this.LeftPath {
		sink.WriteUint8(uint8(item.Height))
		sink.WriteInt64(item.Size)
		sink.WriteInt64(item.Version)
		sink.WriteVarBytes(item.Left)
		sink.WriteVarBytes(item.Right)
	}
	sink.WriteUint32(uint32(len(this.InnerNodes)))
	for _, item := range this.InnerNodes {
		sink.WriteUint32(uint32(len(item)))
		for _, item1 := range item {
			sink.WriteUint8(uint8(item1.Height))
			sink.WriteInt64(item1.Size)
			sink.WriteInt64(item1.Version)
			sink.WriteVarBytes(item1.Left)
			sink.WriteVarBytes(item1.Right)
		}
	}
	sink.WriteUint32(uint32(len(this.Leaves)))
	for _, item := range this.Leaves {
		sink.WriteVarBytes(item.Key)
		sink.WriteVarBytes(item.ValueHash)
		sink.WriteInt64(item.Version)
	}
}

func (this *StoreProof) Deserialization(source *common.ZeroCopySource) error {
	leftPathLen, _ := source.NextUint32()
	this.LeftPath = make([]iavl.ProofInnerNode, leftPathLen)
	for _, item := range this.LeftPath {
		height, _ := source.NextUint8()
		item.Height = int8(height)
		item.Size, _ = source.NextInt64()
		item.Version, _ = source.NextInt64()
		item.Left, _ = source.ReadVarBytes()
		item.Right, _ = source.ReadVarBytes()
	}
	innerNodesLen, _ := source.NextUint32()
	this.InnerNodes = make([]iavl.PathToLeaf, innerNodesLen)
	for i, _ := range this.InnerNodes {
		pathToLeafLen, _ := source.NextUint32()
		this.InnerNodes[i] = make([]iavl.ProofInnerNode, pathToLeafLen)
		for _, item := range this.InnerNodes[i] {
			height, _ := source.NextUint8()
			item.Height = int8(height)
			item.Size, _ = source.NextInt64()
			item.Version, _ = source.NextInt64()
			item.Left, _ = source.ReadVarBytes()
			item.Right, _ = source.ReadVarBytes()
		}
	}
	leavesLen, _ := source.NextUint32()
	this.Leaves = make([]iavl.ProofLeafNode, leavesLen)
	for _, item := range this.Leaves {
		item.Key, _ = source.ReadVarBytes()
		item.ValueHash, _ = source.ReadVarBytes()
		item.Version, _ = source.NextInt64()
	}
	return nil
}
