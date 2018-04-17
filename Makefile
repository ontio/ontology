GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
Minversion := $(shell date)
BUILD_NODE_PAR = -ldflags "-X github.com/ontio/ontology/common/config.Version=$(VERSION)" #-race
BUILD_NODECTL_PAR = -ldflags "-X main.Version=$(VERSION)"

ARCH=$(shell uname -m)
DBUILD=docker build
DOCKER_NS ?= ontio
DOCKER_TAG=$(ARCH)-$(VERSION)
CONFIG_FILES=$(shell ls *.json)
WALLET_FILE=wallet.dat

all: ontology nodectl

ontology:
	$(GC)  $(BUILD_NODE_PAR) -o ontology main.go

nodectl:
	$(GC)  $(BUILD_NODECTL_PAR) nodectl.go

format:
	$(GOFMT) -w main.go

$(WALLET_FILE): 
	./nodectl wallet -c -p passwordtest -n $(WALLET_FILE) 

docker/payload: Makefile ontology nodectl docker/Dockerfile $(CONFIG_FILES) $(WALLET_FILE)
	mkdir -p $@
	cp docker/Dockerfile $@
	cp ontology $@
	cp nodectl $@
	tar czf $@/config.tgz $(CONFIG_FILES) $(WALLET_FILE)

docker: docker/payload docker/Dockerfile 
	@echo "Building docker"
	$(DBUILD) -t $(DOCKER_NS)/ontology docker/payload
	docker tag $(DOCKER_NS)/ontology $(DOCKER_NS)/ontology:$(DOCKER_TAG)
	@touch $@

clean:
	rm -rf *.8 *.o *.out *.6
	rm -rf ontology nodectl docker/payload

