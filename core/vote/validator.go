package vote

import (
	//"math"
	"sort"

	. "github.com/Ontology/common"
	//"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
)

func GetValidators(txs []*types.Transaction) ([]*crypto.PubKey, error) {
	// TODO implement vote
	return genesis.GenesisBookKeepers, nil
}

/*
func GetValidators(txs []*types.Transaction) ([]*crypto.PubKey, error) {
	votes, validators, err := ledger.DefaultLedger.Store.GetVotesAndEnrollments(txs)
	if err != nil {
		return nil, err
	}
	validatorCount := int(math.Max(float64(weightedAverage(votes)), float64(len(ledger.StandbyBookKeepers))))
	if err != nil {
		return nil, err
	}
	result := make(map[string]Fixed64)
	for _, v := range validators {
		key, _ := v.EncodePoint(false)
		result[string(key)] = Fixed64(0)
	}
	for _, v := range votes {
		count := int(math.Min(float64(len(v.PublicKeys)), float64(validatorCount)))
		for i := 0; i < count; i++ {
			if crypto.ContainPubKey(v.PublicKeys[i], validators) >= 0 {

				key, _ := v.PublicKeys[i].EncodePoint(false)
				result[string(key)] += Fixed64(v.Count.GetData() / int64(count))
			}
		}
	}

	values := sortMapByValue(result)
	values = values[:validatorCount]
	keys := make(crypto.PubKeySlice, len(values))
	for i, k := range values {
		key, _ := crypto.DecodePoint([]byte(k))
		keys[i] = key
	}
	sort.Sort(keys)
	return keys, nil
}
*/

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
func sortMapByValue(m map[string]Fixed64) []string {
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
