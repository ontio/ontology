package vote

import (
	"sort"

	"github.com/Ontology/common"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/types"
	"github.com/ontio/ontology-crypto/keypair"
)

func GetValidators(txs []*types.Transaction) ([]keypair.PublicKey, error) {
	// TODO implement vote
	return genesis.GenesisBookkeepers, nil
}

func weightedAverage(votes []*states.VoteState) int64 {
	var sumWeight, sumValue int64
	for _, v := range votes {
		sumWeight += v.Count.GetData()
		sumValue += v.Count.GetData() * int64(len(v.PublicKeys))
	}
	if sumValue == 0 {
		return 0
	}
	return sumValue / sumWeight
}

type Pair struct {
	Key   string
	Value int64
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int      { return len(p) }
func (p PairList) Less(i, j int) bool {
	if p[j].Value < p[i].Value {
		return true
	} else if p[j].Value > p[i].Value {
		return false
	} else {
		return p[j].Key < p[i].Key
	}
}

// A function to turn a map into a PairList, then sort and return it.
func sortMapByValue(m map[string]common.Fixed64) []string {
	p := make(PairList, 0, len(m))
	for k, v := range m {
		p = append(p, Pair{k, v.GetData()})
	}
	sort.Sort(p)
	keys := make([]string, len(m))
	for i, k := range p {
		keys[i] = k.Key
	}
	return keys
}
