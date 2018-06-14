GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
BUILD_NODE_PAR = -ldflags "-X github.com/ontio/ontology/common/config.Version=$(VERSION)" #-race

ARCH=$(shell uname -m)
DBUILD=docker build
DRUN=docker run
DOCKER_NS ?= ontio
DOCKER_TAG=$(ARCH)-$(VERSION)
ONT_CFG_IN_DOCKER=config-solo.json
WALLET_FILE=wallet.dat

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)
TOOLS=./tools
NATIVE_ABI=$(TOOLS)/abi/native

ontology: $(SRC_FILES)
	$(GC)  $(BUILD_NODE_PAR) -o ontology main.go

tools: $(SRC_FILES)
	$(GC)  $(BUILD_NODE_PAR) -o sigsvr sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir $(TOOLS) ;fi
	@mv sigsvr $(TOOLS)
	@if [ ! -d $(NATIVE_ABI) ];then mkdir -p $(NATIVE_ABI) ;fi
	@cp ./cmd/abi/native/*.json $(NATIVE_ABI)

all: ontology tools
	
format:
	$(GOFMT) -w main.go

$(WALLET_FILE):
	@if [ ! -e $(WALLET_FILE) ]; then $(error Please create wallet file first) ; fi

docker/payload: docker/build/bin/ontology docker/Dockerfile $(ONT_CFG_IN_DOCKER) $(WALLET_FILE)
	@echo "Building ontology payload"
	@mkdir -p $@
	@cp docker/Dockerfile $@
	@cp docker/build/bin/ontology $@
	@cp -f $(ONT_CFG_IN_DOCKER) $@/config.json
	@cp -f $(WALLET_FILE) $@
	@tar czf $@/config.tgz -C $@ config.json $(WALLET_FILE)
	@touch $@

docker/build/bin/%: Makefile
	@echo "Building ontology in docker"
	@mkdir -p docker/build/bin docker/build/pkg
	@$(DRUN) --rm \
		-v $(abspath docker/build/bin):/go/bin \
		-v $(abspath docker/build/pkg):/go/pkg \
		-v $(GOPATH)/src:/go/src \
		-w /go/src/github.com/ontio/ontology \
		golang:1.9.5-stretch \
		$(GC)  $(BUILD_NODE_PAR) -o docker/build/bin/ontology main.go
	@touch $@

docker: Makefile docker/payload docker/Dockerfile 
	@echo "Building ontology docker"
	@$(DBUILD) -t $(DOCKER_NS)/ontology docker/payload
	@docker tag $(DOCKER_NS)/ontology $(DOCKER_NS)/ontology:$(DOCKER_TAG)
	@touch $@

clean:
	rm -rf *.8 *.o *.out *.6
	rm -rf ontology tools docker/payload docker/build

