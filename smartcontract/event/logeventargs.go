package event

import (
	"github.com/Ontology/vm/neovm/interfaces"
	"github.com/Ontology/common"
)

type LogEventArgs struct {
	container interfaces.ICodeContainer
	codeHash  common.Uint160
	message   string
}