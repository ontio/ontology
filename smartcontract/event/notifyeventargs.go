package event

import (
	"github.com/Ontology/vm/neovm/interfaces"
	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/types"
)

type NotifyEventArgs struct {
	container interfaces.ICodeContainer
	codeHash  common.Uint160
	state     types.StackItemInterface
}

