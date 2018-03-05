package httprestful

import (
	"github.com/Ontology/http/httprestful/common"
	. "github.com/Ontology/http/httprestful/restful"
	. "github.com/Ontology/net/protocol"
)

func StartServer(n Noder) {
	common.SetNode(n)
	func() {
		rest := InitRestServer()
		go rest.Start()
	}()
}

