/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package graphql

import (
	"net"
	"net/http"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/ontio/ontology/v2/common"
	"github.com/ontio/ontology/v2/common/config"
	"github.com/ontio/ontology/v2/common/log"
	"github.com/ontio/ontology/v2/core/payload"
	"github.com/ontio/ontology/v2/core/types"
	"github.com/ontio/ontology/v2/http/base/actor"
	comm "github.com/ontio/ontology/v2/http/base/common"
	"github.com/ontio/ontology/v2/http/graphql/schema"
	"github.com/ontio/ontology/v2/smartcontract/service/native/utils"
	"golang.org/x/net/netutil"
)

var ontSchema *graphql.Schema

func init() {
	resolver := &resolver{}
	s, err := schema.Asset("schema.graphql")
	if err != nil {
		panic(err)
	}

	ontSchema = graphql.MustParseSchema(string(s), resolver, graphql.UseFieldResolvers())
}

type resolver struct{}
type block struct {
	Header       *header
	Transactions []*transaction
}

type header struct {
	Version        Uint32
	Hash           H256
	PrevHash       H256
	Height         Uint32
	Timestamp      Uint32
	BlockRoot      H256
	TxsRoot        H256
	ConsensusData  Uint64
	NextBookkeeper Addr
	Bookkeepers    []PubKey
	SigData        []string
}

type invokeCodePayload struct {
	Code string
}

type deployCodePayload struct {
	Code    string
	VmType  string
	Name    string
	Version string
	Author  string
	Email   string
	Desc    string
}

type TxPayload struct {
	pl interface{}
}

func (self *TxPayload) ToInvokeCode() (*invokeCodePayload, bool) {
	pl, ok := self.pl.(*invokeCodePayload)
	return pl, ok
}

func (self *TxPayload) ToDeployCode() (*deployCodePayload, bool) {
	pl, ok := self.pl.(*deployCodePayload)
	return pl, ok
}

func (self *TxPayload) Code() string {
	switch pd := self.pl.(type) {
	case *invokeCodePayload:
		return pd.Code
	case *deployCodePayload:
		return pd.Code
	default:
		panic("unreachable")
	}
}

func NewTxPayload(pl types.Payload) *TxPayload {
	switch val := pl.(type) {
	case *payload.InvokeCode:
		return &TxPayload{pl: &invokeCodePayload{Code: common.ToHexString(val.Code)}}
	case *payload.DeployCode:
		vmty := "Neo"
		if val.VmType() == payload.WASMVM_TYPE {
			vmty = "Wasm"
		}
		dp := &deployCodePayload{
			Code:    common.ToHexString(val.GetRawCode()),
			VmType:  vmty,
			Name:    val.Name,
			Version: val.Version,
			Author:  val.Author,
			Email:   val.Email,
			Desc:    val.Description,
		}
		return &TxPayload{pl: dp}
	default:
		panic("unreachable")
	}
}

func NewTransaction(tx *types.Transaction, height uint32) *transaction {
	ty := convTxType(tx)
	var sigs []*Sig
	for _, val := range tx.Sigs {
		sig, err := val.GetSig()
		if err == nil {
			var sigdata []string
			for _, data := range sig.SigData {
				sigdata = append(sigdata, common.ToHexString(data))
			}
			var pubkey []PubKey
			for _, val := range sig.PubKeys {
				pubkey = append(pubkey, PubKey(common.PubKeyToHex(val)))
			}
			sigs = append(sigs, &Sig{
				SigData: sigdata,
				PubKeys: pubkey,
				M:       Uint32(sig.M),
			})
		}
	}
	t := &transaction{
		Version:  Uint32(tx.Version),
		Hash:     H256(tx.Hash()),
		Nonce:    Uint32(tx.Nonce),
		TxType:   ty,
		GasPrice: Uint32(tx.GasPrice),
		GasLimit: Uint32(tx.GasLimit),
		Payer:    Addr{tx.Payer},
		Payload:  NewTxPayload(tx.Payload),
		Sigs:     sigs,
		Height:   Uint32(height),
	}

	return t
}

func NewBlock(b *types.Block) *block {
	var txs []*transaction
	for _, tx := range b.Transactions {
		txs = append(txs, NewTransaction(tx, b.Header.Height))
	}
	return &block{
		Header:       NewHeader(b.Header),
		Transactions: txs,
	}
}

func NewHeader(h *types.Header) *header {
	var pubKeys []PubKey
	for _, k := range h.Bookkeepers {
		pubKeys = append(pubKeys, PubKey(common.PubKeyToHex(k)))
	}
	var sigData []string
	for _, sig := range h.SigData {
		sigData = append(sigData, common.ToHexString(sig))
	}

	hd := &header{
		Version:        Uint32(h.Version),
		Hash:           H256(h.Hash()),
		PrevHash:       H256(h.PrevBlockHash),
		Height:         Uint32(h.Height),
		Timestamp:      Uint32(h.Timestamp),
		BlockRoot:      H256(h.BlockRoot),
		TxsRoot:        H256(h.TransactionsRoot),
		ConsensusData:  Uint64(h.ConsensusData),
		NextBookkeeper: Addr{h.NextBookkeeper},
		Bookkeepers:    pubKeys,
		SigData:        sigData,
	}

	return hd
}

func (self *resolver) GetBlockByHeight(args struct{ Height Uint32 }) (*block, error) {
	b, err := actor.GetBlockByHeight(uint32(args.Height))
	if err != nil {
		return nil, err
	}

	return NewBlock(b), nil
}

func (self *resolver) GetBlockByHash(args struct{ Hash H256 }) (*block, error) {
	b, err := actor.GetBlockFromStore(common.Uint256(args.Hash))
	if err != nil {
		return nil, err
	}

	return NewBlock(b), nil
}

func (self *resolver) GetBlockHash(args struct{ Height Uint32 }) H256 {
	return H256(actor.GetBlockHashFromStore(uint32(args.Height)))
}

type balance struct {
	Ont    Uint64
	Ong    Uint64
	Height Uint32
}

type TxType string

const INVOKE_NEO TxType = "INVOKE_NEO"
const INVOKE_WASM TxType = "INVOKE_WASM"
const DEPLOY_NEO TxType = "DEPLOY_NEO"
const DEPLOY_WASM TxType = "DEPLOY_WASM"

func convTxType(tx *types.Transaction) TxType {
	switch pl := tx.Payload.(type) {
	case *payload.InvokeCode:
		if tx.TxType == types.InvokeNeo {
			return INVOKE_NEO
		} else {
			return INVOKE_WASM
		}
	case *payload.DeployCode:
		switch pl.VmType() {
		case payload.NEOVM_TYPE:
			return DEPLOY_NEO
		case payload.WASMVM_TYPE:
			return DEPLOY_WASM
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

type transaction struct {
	Version  Uint32
	Hash     H256
	Nonce    Uint32
	TxType   TxType
	GasPrice Uint32
	GasLimit Uint32
	Payer    Addr
	Payload  *TxPayload
	Sigs     []*Sig
	Height   Uint32
}

type Sig struct {
	SigData []string
	PubKeys []PubKey
	M       Uint32
}

func (self *resolver) GetTx(args struct{ Hash H256 }) (*transaction, error) {
	height, tx, err := actor.GetTxnWithHeightByTxHash(common.Uint256(args.Hash))
	if err != nil {
		return nil, err
	}

	return NewTransaction(tx, height), nil
}

func (self *resolver) GetBalance(args struct{ Addr Addr }) (*balance, error) {
	balances, height, err := comm.GetContractBalance(0,
		[]common.Address{utils.OntContractAddress, utils.OngContractAddress}, args.Addr.Address, true)
	if err != nil {
		return nil, err
	}

	return &balance{
		Height: Uint32(height),
		Ont:    Uint64(balances[0]),
		Ong:    Uint64(balances[1]),
	}, nil
}

func StartServer(cfg *config.GraphQLConfig) {
	if !cfg.EnableGraphQL || cfg.GraphQLPort == 0 {
		return
	}

	serverMut := http.NewServeMux()
	serverMut.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))

	serverMut.Handle("/query", &relay.Handler{Schema: ontSchema})

	server := &http.Server{Handler: serverMut}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(cfg.GraphQLPort)))
	if err != nil {
		log.Error("start graphql server error: %s", err)
		return
	}
	if cfg.MaxConnections > 0 {
		listener = netutil.LimitListener(listener, int(cfg.MaxConnections))
	}

	log.Infof("start GraphQL service on %d", cfg.GraphQLPort)
	log.Error(server.Serve(listener))
}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.css" rel="stylesheet" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/es6-promise/4.1.1/es6-promise.auto.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/2.0.3/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/16.2.0/umd/react.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react-dom/16.2.0/umd/react-dom.production.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.min.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/query", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
