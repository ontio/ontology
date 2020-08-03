package protocols

import (
	"github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rpc"
	"github.com/ontio/ontology/p2pserver/protocols/subnet"
)

func RegisterProposeOfflineVote(subnet *subnet.SubNet) {
	// curl http://localhost:20337/local -v -d '{"method":"proposeOfflineVote", "params":["pubkey1", "pubkey2"]}'
	rpc.HandleFunc("proposeOfflineVote", func(params []interface{}) map[string]interface{} {
		var nodes []string
		for _, key := range params {
			switch pubKey := key.(type) {
			case string:
				nodes = append(nodes, pubKey)
			default:
				return rpc.ResponsePack(error.INVALID_PARAMS, "")
			}
		}

		err := subnet.ProposeOffline(nodes)
		if err != nil {
			return rpc.ResponsePack(error.INTERNAL_ERROR, err.Error())
		}

		return rpc.ResponseSuccess(nil)
	})

	// curl http://localhost:20337/local -v -d '{"method":"getOfflineVotes", "params":[]}'
	rpc.HandleFunc("getOfflineVotes", func(params []interface{}) map[string]interface{} {
		votes := subnet.GetOfflineVotes()

		return rpc.ResponseSuccess(votes)
	})
}
