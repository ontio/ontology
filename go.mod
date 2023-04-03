module github.com/ontio/ontology

go 1.17

require (
	github.com/JohnCGriffin/overflow v0.0.0-20170615021017-4d914c927216
	github.com/blang/semver v3.5.1+incompatible
	github.com/ethereum/go-ethereum v1.10.22
	github.com/gammazero/workerpool v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uiprogress v0.0.1
	github.com/graph-gophers/graphql-go v1.3.0
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/holiman/uint256 v1.2.0
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef
	github.com/itchyny/base58-go v0.1.0
	github.com/laizy/bigint v0.1.3
	github.com/ontio/ontology-crypto v1.2.1
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/ontio/wagon v0.4.2
	github.com/pborman/uuid v1.2.0
	github.com/prometheus/client_golang v0.9.1
	github.com/scylladb/go-set v1.0.2
	github.com/stretchr/testify v1.7.2
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/urfave/cli v1.22.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net v0.0.0-20220607020251-c690dde0001d
)

require (
	github.com/VictoriaMetrics/fastcache v1.6.0 // indirect
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gosuri/uilive v0.0.3 // indirect
	github.com/hashicorp/go-bexpr v0.1.10 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/huin/goupnp v1.0.3 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/mitchellh/pointerstructure v1.2.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910 // indirect
	github.com/prometheus/common v0.0.0-20181113130724-41aa239b4cce // indirect
	github.com/prometheus/procfs v0.0.0-20181005140218-185b4288413d // indirect
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/urfave/cli/v2 v2.10.2 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net => github.com/golang/net v0.0.0-20210924151903-3ad01bbaa167
	golang.org/x/sys => github.com/golang/sys v0.0.0-20210927052749-1cf2251ac284
	golang.org/x/text => github.com/golang/text v0.3.0
)
