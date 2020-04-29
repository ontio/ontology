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

package pdp

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ontio/ontology/smartcontract/service/native/ontfs/pdp/merkle_pdp"
	"github.com/ontio/ontology/smartcontract/service/native/ontfs/pdp/types"
)

const (
	MerklePdp = 1
)

const (
	MaxMerklePdpChallengeBlock = 6
)

type Pdp struct {
	Version uint64
}

func NewPdp(version uint64) *Pdp {
	return &Pdp{Version: version}
}

func (p *Pdp) GenUniqueIdWithFileBlocks(fileBlocks []types.Block) ([]byte, error) {
	var uniqueId []byte
	uniqueIdPrefix := make([]byte, 8)
	switch p.Version {
	case MerklePdp:
		binary.LittleEndian.PutUint64(uniqueIdPrefix, MerklePdp)
		uniqueId = append(uniqueId, uniqueIdPrefix...)
		for _, block := range fileBlocks {
			rootHash := merkle_pdp.CalcRootHash(block[:])
			uniqueId = append(uniqueId, rootHash...)
		}
		return uniqueId, nil
	default:
		return nil, fmt.Errorf("GenUniqueIdWithFileBlocks pdpVersion error")
	}
}

//GenChallenge compute the index to choose block
func (p *Pdp) GenChallenge(nodeId [20]byte, blockHash []byte, fileBlockNum uint64) ([]uint64, error) {
	var challengeNum uint64
	var challenge []uint64

	switch p.Version {
	case MerklePdp:
		if fileBlockNum > MaxMerklePdpChallengeBlock {
			challengeNum = MaxMerklePdpChallengeBlock
		} else {
			challengeNum = fileBlockNum
		}
		challenge = make([]uint64, challengeNum)

		blockNum := big.NewInt(int64(fileBlockNum))
		plant := append(nodeId[:], blockHash...)
		vAdded := make([]byte, 8)
		for i := uint64(0); i < challengeNum; i++ {
			binary.LittleEndian.PutUint64(vAdded, i)
			hash := sha256.Sum256(append(plant, vAdded...))
			bigTmp := new(big.Int).SetBytes(hash[:])
			challenge[i] = bigTmp.Mod(bigTmp, blockNum).Uint64()
		}
		return challenge, nil
	default:
		return nil, fmt.Errorf("GenChallenge pdpVersion error")
	}
}

//BuildProof need parameters
func (p *Pdp) GenProofWithBlocks(fileBlocks []types.Block, uniqueId []byte, challenge []uint64) ([]byte, error) {
	var proof []byte
	proofPrefix := make([]byte, 8)

	switch p.Version {
	case MerklePdp:
		binary.LittleEndian.PutUint64(proofPrefix, MerklePdp)
		proof = append(proof, proofPrefix...)
		for _, chl := range challenge {
			blockProof := merkle_pdp.MerkleProof(fileBlocks[chl], 0)
			merkleProofData, err := json.Marshal(blockProof)
			if err != nil {
				return nil, err
			}
			proof = append(proof, merkleProofData...)
		}
		return proof, nil
	default:
		return nil, fmt.Errorf("GenProofWithBlocks pdpVersion error")
	}
}

//GetPdpVersionFromUniqueId get pdp version from uniqueId data
func GetPdpVersionFromUniqueId(uniqueId []byte) uint64 {
	uniqueIdPrefix := uniqueId[0:8]
	return binary.LittleEndian.Uint64(uniqueIdPrefix)
}

//GetPdpVersionFromProof get pdp version from proof data
func GetPdpVersionFromProof(proof []byte) uint64 {
	proofPrefix := proof[0:8]
	return binary.LittleEndian.Uint64(proofPrefix)
}

//VerifyProof used in consensus algorithm
func VerifyProofWithUniqueId(uniqueId []byte, proof []byte, challenge []uint64) error {
	proofPrefix := proof[0:8]
	uniqueIdPrefix := uniqueId[0:8]
	proofPdpVersion := binary.LittleEndian.Uint64(proofPrefix)
	uniqueIdPdpVersion := binary.LittleEndian.Uint64(uniqueIdPrefix)
	if proofPdpVersion != uniqueIdPdpVersion {
		return fmt.Errorf("[VerifyProofWithUniqueId] pdpVersion is not match")
	}
	switch proofPdpVersion {
	case MerklePdp:
		challengeCount := len(challenge)

		blocksId := uniqueId[8:]
		blocksIdLen := len(blocksId)
		if blocksIdLen%merkle_pdp.MerkleRootLen != 0 {
			return fmt.Errorf("[VerifyProofWithUniqueId] uniqueId length error")
		}

		merkleProofData := proof[8:]
		merkleProofDataLen := len(merkleProofData)

		if merkleProofDataLen%challengeCount != 0 {
			return fmt.Errorf("[VerifyProofWithUniqueId] proof length error")
		}
		singleProofLen := merkleProofDataLen / challengeCount
		for index, chl := range challenge {
			var merkleProof [][]byte
			blockProof := merkleProofData[index*singleProofLen : (index+1)*singleProofLen]
			if err := json.Unmarshal(blockProof, &merkleProof); err != nil {
				return fmt.Errorf("[VerifyProofWithUniqueId] error: %s", err.Error())
			}

			blockId := blocksId[chl*merkle_pdp.MerkleRootLen : (chl+1)*merkle_pdp.MerkleRootLen]
			return merkle_pdp.VerifyMerkleProof(merkleProof, blockId, 0)
		}
	default:
		return fmt.Errorf("[VerifyProofWithUniqueId] pdpVersion error")
	}
	return nil
}
