package restful

import (
	. "github.com/Ontology/http/base/rest"
	. "github.com/Ontology/http/restful/restful"
	. "github.com/Ontology/net/protocol"
)

func StartServer(n Noder) {
	SetNode(n)
	func() {
		rt := InitRestServer()
		go rt.Start()
	}()
}

