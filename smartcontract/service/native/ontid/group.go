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
package ontid

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const MAX_DEPTH = 8

// Group defines a group control logic
type Group struct {
	Members   []interface{} `json:"members"`
	Threshold uint          `json:"threshold"`
}

type GroupJson struct {
	Members   []interface{} `json:"members"`
	Threshold uint          `json:"threshold"`
}

func parse(g *Group) *GroupJson {
	gr := &GroupJson{
		Members:   make([]interface{}, len(g.Members)),
		Threshold: g.Threshold,
	}
	for i := 0; i < len(g.Members); i++ {
		switch t := g.Members[i].(type) {
		case []byte:
			gr.Members[i] = string(t)
		case *Group:
			gr.Members[i] = parse(t)
		default:
			panic("invalid member type")
		}
	}
	return gr
}

func (g *Group) ToJson() []byte {
	j, _ := json.Marshal(parse(g))
	return j
}

func rDeserialize(data []byte, depth uint) (*Group, error) {
	if depth == MAX_DEPTH {
		return nil, fmt.Errorf("recursion is too deep")
	}

	g := Group{}
	buf := common.NewZeroCopySource(data)

	// parse members
	num, err := utils.DecodeVarUint(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing number: %s", err)
	}

	for i := uint64(0); i < num; i++ {
		m, err := utils.DecodeVarBytes(buf)
		if err != nil {
			return nil, fmt.Errorf("error parsing group members: %s", err)
		}
		if len(m) > 8 && bytes.Equal(m[:8], []byte("did:ont:")) {
			g.Members = append(g.Members, m)
		} else {
			// parse recursively
			g1, err := rDeserialize(m, depth+1)
			if err != nil {
				return nil, fmt.Errorf("error parsing subgroup: %s", err)
			}
			g.Members = append(g.Members, g1)
		}
	}

	// parse threshold
	t, err := utils.DecodeVarUint(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing group threshold: %s", err)
	}
	if t > uint64(len(g.Members)) {
		return nil, fmt.Errorf("invalid threshold")
	}

	g.Threshold = uint(t)

	return &g, nil
}

func deserializeGroup(data []byte) (*Group, error) {
	return rDeserialize(data, 0)
}

func validateMembers(srvc *native.NativeService, g *Group) error {
	for _, m := range g.Members {
		switch t := m.(type) {
		case []byte:
			key, err := encodeID(t)
			if err != nil {
				return fmt.Errorf("invalid id: %s", string(t))
			}
			// ID must exists
			if !isValid(srvc, key) {
				return fmt.Errorf("id %s not registered", string(t))
			}
			// Group member must have its own public key
			pk, err := getPk(srvc, key, 1)
			if err != nil || pk == nil {
				return fmt.Errorf("id %s has no public keys", string(t))
			}
		case *Group:
			if err := validateMembers(srvc, t); err != nil {
				return err
			}
		default:
			panic("group member type error")
		}
	}
	return nil
}

type Signer struct {
	Id    []byte
	Index uint32
}

func SerializeSigners(s []Signer) []byte {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarUint(sink, uint64(len(s)))
	for _, v := range s {
		sink.WriteVarBytes(v.Id)
		utils.EncodeVarUint(sink, uint64(v.Index))
	}
	return sink.Bytes()
}

func deserializeSigners(data []byte) ([]Signer, error) {
	buf := common.NewZeroCopySource(data)
	num, err := utils.DecodeVarUint(buf)
	if err != nil {
		return nil, err
	}

	signers := []Signer{}
	for i := uint64(0); i < num; i++ {
		id, err := utils.DecodeVarBytes(buf)
		if err != nil {
			return nil, err
		}
		index, err := utils.DecodeVarUint(buf)
		if err != nil {
			return nil, err
		}

		signer := Signer{id, uint32(index)}
		signers = append(signers, signer)
	}

	return signers, nil
}

func findSigner(id []byte, signers []Signer) bool {
	for _, signer := range signers {
		if bytes.Equal(signer.Id, id) {
			return true
		}
	}
	return false
}

func verifyThreshold(g *Group, signers []Signer) bool {
	var signed uint = 0
	for _, member := range g.Members {
		switch t := member.(type) {
		case []byte:
			if findSigner(t, signers) {
				signed += 1
			}
		case *Group:
			if verifyThreshold(t, signers) {
				signed += 1
			}
		default:
			panic("invalid group member type")
		}
	}
	return signed >= g.Threshold
}

func verifyGroupSignature(srvc *native.NativeService, g *Group, signers []Signer) bool {
	if !verifyThreshold(g, signers) {
		return false
	}

	for _, signer := range signers {
		key, err := encodeID(signer.Id)
		if err != nil {
			return false
		}
		if checkWitnessByIndex(srvc, key, signer.Index) != nil {
			return false
		}
	}
	return true
}
