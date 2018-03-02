package httprestful

import (
	"github.com/Ontology/net/httprestful/common"
	. "github.com/Ontology/net/httprestful/restful"
	. "github.com/Ontology/net/protocol"
)

func StartServer(n Noder) {
	common.SetNode(n)
	func() {
		rest := InitRestServer()
		go rest.Start()
	}()
}

