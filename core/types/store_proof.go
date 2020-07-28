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

package types

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
	for i := 0;i < int(leftPathLen);i ++ {
		height, _ := source.NextUint8()
		this.LeftPath[i].Height = int8(height)
		this.LeftPath[i].Size, _ = source.NextInt64()
		this.LeftPath[i].Version, _ = source.NextInt64()
		this.LeftPath[i].Left, _ = source.ReadVarBytes()
		this.LeftPath[i].Right, _ = source.ReadVarBytes()
	}
	innerNodesLen, _ := source.NextUint32()
	this.InnerNodes = make([]iavl.PathToLeaf, innerNodesLen)
	for i, _ := range this.InnerNodes {
		pathToLeafLen, _ := source.NextUint32()
		this.InnerNodes[i] = make([]iavl.ProofInnerNode, pathToLeafLen)
		for j := 0;j < int(pathToLeafLen);j ++ {
			height, _ := source.NextUint8()
			this.InnerNodes[i][j].Height = int8(height)
			this.InnerNodes[i][j].Size, _ = source.NextInt64()
			this.InnerNodes[i][j].Version, _ = source.NextInt64()
			this.InnerNodes[i][j].Left, _ = source.ReadVarBytes()
			this.InnerNodes[i][j].Right, _ = source.ReadVarBytes()
		}
	}
	leavesLen, _ := source.NextUint32()
	this.Leaves = make([]iavl.ProofLeafNode, leavesLen)
	for i := 0;i < int(leavesLen);i ++ {
		this.Leaves[i].Key, _ = source.ReadVarBytes()
		this.Leaves[i].ValueHash, _ = source.ReadVarBytes()
		this.Leaves[i].Version, _ = source.NextInt64()
	}
	return nil
}
