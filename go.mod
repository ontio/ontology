module github.com/ontio/ontology

go 1.17

require (
	github.com/JohnCGriffin/overflow v0.0.0-20170615021017-4d914c927216
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gammazero/workerpool v1.1.2
	github.com/gorilla/websocket v1.4.1
	github.com/gosuri/uilive v0.0.3 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/graph-gophers/graphql-go v1.2.1-0.20210916100229-446a2dd13dd5
	github.com/hashicorp/golang-lru v0.5.4
	github.com/holiman/uint256 v1.1.1
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef
	github.com/itchyny/base58-go v0.1.0
	github.com/json-iterator/go v1.1.10
	github.com/laizy/bigint v0.1.3
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/ontio/ontology-crypto v1.2.1
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/ontio/wagon v0.4.2
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc // indirect
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v0.9.1
	github.com/scylladb/go-set v1.0.2
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/urfave/cli v1.22.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net => github.com/golang/net v0.0.0-20210924151903-3ad01bbaa167
	golang.org/x/sys => github.com/golang/sys v0.0.0-20210927052749-1cf2251ac284
	golang.org/x/text => github.com/golang/text v0.3.0
)
