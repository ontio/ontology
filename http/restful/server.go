package restful

import (
	"github.com/Ontology/http/restful/common"
	. "github.com/Ontology/http/restful/restful"
	. "github.com/Ontology/net/protocol"
)

func StartServer(n Noder) {
	common.SetNode(n)
	func() {
		rest := InitRestServer()
		go rest.Start()
	}()
}

