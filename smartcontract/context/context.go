package context

import (
	"github.com/Ontology/common"
	"github.com/Ontology/smartcontract/event"
	vmtypes "github.com/Ontology/vm/types"
)

type ContextRef interface {
	PushContext(context *Context)
	CurrentContext() *Context
	CallingContext() *Context
	EntryContext() *Context
	PopContext()
	CheckWitness(address common.Address) bool
	PushNotifications(notifications []*event.NotifyEventInfo)
	Execute() error
}

type Context struct {
	ContractAddress common.Address
	Code            vmtypes.VmCode
}
