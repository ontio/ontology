package bgn

import (
	"errors"
	"math"
	"math/big"
	"sync"

	"github.com/Nik-U/pbc"
)

var tableG1 sync.Map
var tableGT sync.Map

var usingCache = false
var tablesComputed = false

func computeTableG1(gen *pbc.Element, bound int64) {

	aux := gen.NewFieldElement()
	aux.Set(gen)

	for j := int64(0); j <= bound; j++ {
		tableG1.Store(aux.String(), j)
		aux.Mul(aux, gen)
	}
}

func computeTableGT(gen *pbc.Element, bound int64) {

	aux := gen.NewFieldElement()
	aux.Set(gen)

	for j := int64(0); j <= bound; j++ {
		tableGT.Store(aux.String(), j)
		aux.Mul(aux, gen)
	}
}

// PrecomputeTables builds the maps necessary
// for the giant step, baby step algorithm
func (pk *PublicKey) PrecomputeTables(genG1 *pbc.Element, genGT *pbc.Element) {

	// sqrt of the largest possible message
	bound := int64(math.Ceil(math.Sqrt(float64(pk.T.Int64())))) + 1

	// pre-compute the tables for the giant steps
	computeTableGT(genGT, bound)
	computeTableG1(genG1, bound)

	tablesComputed = true
}

// obtain the discrete log in O(sqrt(T)) time using giant step baby step algorithm
func (pk *PublicKey) getDL(csk *pbc.Element, gsk *pbc.Element, l2 bool) (*big.Int, error) {

	if !tablesComputed {
		panic("DL tables not computed!")
	}

	bound := int64(math.Ceil(math.Sqrt(float64(pk.T.Int64()))))

	aux := csk.NewFieldElement()

	gamma := gsk.NewFieldElement()
	gamma.Set(gsk)
	gamma.MulBig(gamma, big.NewInt(0))

	aux.Set(csk)
	aux.Mul(aux, gamma)

	gamma.Set(gsk)
	gamma.MulBig(gamma, big.NewInt(bound))

	var val *big.Int
	var found bool

	for i := int64(0); i <= bound; i++ {

		found = false
		val = big.NewInt(0)

		if l2 {
			value, hit := tableGT.Load(aux.String())
			if v, ok := value.(int64); ok {
				val = big.NewInt(v)
				found = hit
			}

		} else {
			value, hit := tableG1.Load(aux.String())
			if v, ok := value.(int64); ok {
				val = big.NewInt(v)
				found = hit
			}
		}

		if found {
			dl := big.NewInt(i*bound + val.Int64() + 1)

			return dl, nil
		}
		aux.Div(aux, gamma)
	}

	return nil, errors.New("cannot find discrete log; out of bounds")
}
