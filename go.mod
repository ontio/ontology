module github.com/ontio/ontology

go 1.12

require (
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.8.23
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/gorilla/websocket v1.2.0
	github.com/gosuri/uilive v0.0.3 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/hashicorp/golang-lru v0.5.3
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/itchyny/base58-go v0.0.5
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/ontio/ontology-crypto v1.0.5
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/orcaman/concurrent-map v0.0.0-20190314100340-2693aad1ed75 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/urfave/cli v1.21.0
	github.com/valyala/bytebufferpool v1.0.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20190412213103-97732733099d
	golang.org/x/text v0.3.0
	golang.org/x/tools v0.0.0-20180221164845-07fd8470d635
)

replace golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190701094942-4def268fd1a4

replace golang.org/x/net => github.com/golang/net v0.0.0-20190724013045-ca1201d0de80

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20190412213103-97732733099d

replace golang.org/x/text => github.com/golang/text v0.3.0

replace golang.org/x/sync => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f

replace golang.org/x/tools => github.com/golang/tools v0.0.0-20180221164845-07fd8470d635
