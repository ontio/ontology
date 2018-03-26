package context

import (
	"github.com/Ontology/common"
	vmtypes "github.com/Ontology/vm/types"
)

type ContextRef interface {
	LoadContext(context *Context)
	CurrentContext() *Context
	CallingContext() *Context
	EntryContext() *Context
	Execute() error
}


type Context struct {
	ContractAddress common.Address
	Code vmtypes.VmCode
}
