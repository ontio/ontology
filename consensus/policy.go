package consensus

import (
	. "github.com/Ontology/common"
)

type Policy struct {
	PolicyLevel PolicyLevel
	List        []Address
}

func NewPolicy() *Policy {
	return &Policy{}
}

func (p *Policy) Refresh() {
	//TODO: Refresh
}

var DefaultPolicy *Policy

func InitPolicy() {
	DefaultPolicy := NewPolicy()
	DefaultPolicy.Refresh()
}
