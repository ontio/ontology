module github.com/ontio/ontology

go 1.12

require (
	github.com/FactomProject/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/JohnCGriffin/overflow v0.0.0-20170615021017-4d914c927216
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.9.6
	github.com/go-interpreter/wagon v0.6.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/websocket v1.4.1
	github.com/gosuri/uilive v0.0.3 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/hashicorp/golang-lru v0.5.3
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/itchyny/base58-go v0.0.5
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/ontio/go-bip32 v0.0.0-20190520025953-d3cea6894a2b // indirect
	github.com/ontio/ontology-crypto v1.0.5
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/ontio/ontology-go-sdk v1.0.9
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/tyler-smith/go-bip39 v1.0.2 // indirect
	github.com/urfave/cli v1.22.1
	github.com/valyala/bytebufferpool v1.0.0
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/net v0.0.0-20191028085509-fe3aa8a45271
)

replace (
	github.com/go-interpreter/wagon => github.com/ontio/wagon v0.3.1-0.20191012103353-ef8d35ecd300
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/net => github.com/golang/net v0.0.0-20191028085509-fe3aa8a45271
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190412213103-97732733099d
	golang.org/x/text => github.com/golang/text v0.3.0
)
