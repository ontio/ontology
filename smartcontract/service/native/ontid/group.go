package ontid

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// Group defines a group control logic
type Group struct {
	members   []interface{}
	threshold uint
}

func deserializeGroup(data []byte) (*Group, error) {
	g := Group{}
	buf := bytes.NewBuffer(data)

	// parse members
	num, err := utils.ReadVarUint(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing group members: %s", err)
	}

	for i := uint64(0); i < num; i++ {
		m, err := serialization.ReadVarBytes(buf)
		if err != nil {
			return nil, fmt.Errorf("error parsing group members: %s", err)
		}
		if bytes.Equal(m[:8], []byte("did:ont:")) {
			g.members = append(g.members, m)
		} else {
			// parse recursively
			g1, err := deserializeGroup(m)
			if err != nil {
				return nil, fmt.Errorf("error parsing group members: %s", err)
			}
			g.members[i] = g1
		}
	}

	// parse threshold
	t, err := utils.ReadVarUint(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing group threshold: %s", err)
	}

	g.threshold = uint(t)

	return &g, nil
}

func validateMembers(srvc *native.NativeService, g *Group) error {
	for _, m := range g.members {
		switch t := m.(type) {
		case []byte:
			key, err := encodeID(t)
			if err != nil {
				return fmt.Errorf("invalid id: %s", string(t))
			}
			// ID must exists
			if !checkIDExistence(srvc, key) {
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
	id    []byte
	index uint32
}

func deserializeSigners(data []byte) ([]Signer, error) {
	buf := bytes.NewBuffer(data)
	num, err := utils.ReadVarUint(buf)
	if err != nil {
		return nil, err
	}

	signers := []Signer{}
	for i := uint64(0); i < num; i++ {
		id, err := serialization.ReadVarBytes(buf)
		if err != nil {
			return nil, err
		}
		index, err := utils.ReadVarUint(buf)
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
		if bytes.Equal(signer.id, id) {
			return true
		}
	}
	return false
}

func verifyThreshold(g *Group, signers []Signer) bool {
	var signed uint = 0
	for _, member := range g.members {
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
	return signed >= g.threshold
}

func verifyGroupSignature(srvc *native.NativeService, g *Group, signers []Signer) bool {
	if !verifyThreshold(g, signers) {
		return false
	}

	for _, signer := range signers {
		key, err := encodeID(signer.id)
		if err != nil {
			return false
		}
		pk, err := getPk(srvc, key, signer.index)
		if err != nil {
			return false
		}
		if pk.revoked {
			return false
		}
		if checkWitness(srvc, pk.key) != nil {
			return false
		}
	}
	return true
}
