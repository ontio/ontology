GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
BUILD_NODE_PAR = -ldflags "-X github.com/ontio/ontology/common/config.Version=$(VERSION)" #-race

ARCH=$(shell uname -m)
DBUILD=docker build
DRUN=docker run
DOCKER_NS ?= ontio
DOCKER_TAG=$(ARCH)-$(VERSION)
CONFIG_FILES=$(shell ls *.json)
WALLET_FILE=wallet.dat

all: ontology

ontology:
	$(GC)  $(BUILD_NODE_PAR) -o ontology main.go

format:
	$(GOFMT) -w main.go

$(WALLET_FILE): nodectl
	./nodectl wallet -c -p passwordtest -n $(WALLET_FILE) 

docker/payload: docker/build/bin/ontology docker/Dockerfile $(CONFIG_FILES) $(WALLET_FILE)
	@echo "Building ontology payload"
	@mkdir -p $@
	@cp docker/Dockerfile $@
	@cp docker/build/bin/ontology $@
	@tar czf $@/config.tgz $(CONFIG_FILES) $(WALLET_FILE)
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
	rm -rf ontology nodectl docker/payload docker/build

