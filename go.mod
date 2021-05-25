module github.com/ontio/ontology

go 1.12

require (
	github.com/JohnCGriffin/overflow v0.0.0-20170615021017-4d914c927216
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.10.1
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uilive v0.0.3 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/graph-gophers/graphql-go v0.0.0-20201113091052-beb923fada29
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c
	github.com/itchyny/base58-go v0.1.0
	github.com/json-iterator/go v1.1.10
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/ontio/ontology-crypto v1.0.9
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/ontio/wagon v0.4.1
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v1.0.0
	github.com/scylladb/go-set v1.0.2
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/urfave/cli v1.22.1
	github.com/valyala/bytebufferpool v1.0.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200822124328-c89045814202
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/net => github.com/golang/net v0.0.0-20191028085509-fe3aa8a45271
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190412213103-97732733099d
	golang.org/x/text => github.com/golang/text v0.3.0
)
