package restful

import (
	. "github.com/Ontology/http/restful/restful"
)

func StartServer() {
	func() {
		rt := InitRestServer()
		go rt.Start()
	}()
}

