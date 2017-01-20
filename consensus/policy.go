package consensus

import (
	. "GoOnchain/common"
)
type Policy struct {
	PolicyLevel PolicyLevel
	List []Uint160
}

func NewPolicy()  *Policy{
	return &Policy{}
}

func (p *Policy) Refresh(){
	//TODO: Refresh
}

var DefaultPolicy *Policy

func InitPolicy(){
	DefaultPolicy := NewPolicy()
	DefaultPolicy.Refresh()
}