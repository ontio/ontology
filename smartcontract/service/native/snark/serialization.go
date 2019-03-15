package snark

import (
	"errors"

	"github.com/kunxian-xia/bn256"
	"github.com/ontio/ontology/common"
)

const fieldSize = 32
const g1Size = 64
const g2Size = 128

// G1 must be encoded in two 32-byte big-endian numbers
func deserializeG1(point *bn256.G1, source *common.ZeroCopySource) error {
	bytes, eof := source.NextBytes(g1Size)
	if eof {
		return errors.New("eof when deserialize G1")
	}
	_, ok := point.Unmarshal(bytes)
	if !ok {
		return errors.New("failed to unmarshal G1 point")
	}
	return nil
}

// bn256.G2 is of the form (x, y) where x and y are elements from Fp2
//  and every element from Fp2 is of the form y + x*u
//  Fp2 is isomorphic to Fp[u]/(u^2 - non_residue)
func deserializeG2(point *bn256.G2, source *common.ZeroCopySource) error {
	bytes, eof := source.NextBytes(g2Size)
	if eof {
		return errors.New("eof when deserialize G2")
	}
	_, ok := point.Unmarshal(bytes)
	if !ok {
		return errors.New("failed to unmarshal G2 point")
	}
	return nil
}
