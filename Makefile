GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --always --tags --long)
BUILD_NODE_PAR = -ldflags "-w -X github.com/ontio/ontology/v2/common/config.Version=$(VERSION)" #-race

ARCH=$(shell uname -m)
DBUILD=docker build
DRUN=docker run
DOCKER_NS ?= ontio
DOCKER_TAG=$(ARCH)-$(VERSION)

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)
TOOLS=./tools
ABI=$(TOOLS)/abi
NATIVE_ABI_SCRIPT=./cmd/abi/native_abi_script

ontology: $(SRC_FILES)
	CGO_ENABLED=1 $(GC)  $(BUILD_NODE_PAR) -o ontology main.go
 
sigsvr: $(SRC_FILES) abi 
	$(GC)  $(BUILD_NODE_PAR) -o sigsvr cmd-tools/sigsvr/sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr $(TOOLS)

abi: 
	@if [ ! -d $(ABI) ];then mkdir -p $(ABI) ;fi
	@cp $(NATIVE_ABI_SCRIPT)/*.json $(ABI)

tools: sigsvr abi

all: ontology tools

ontology-cross: ontology-windows ontology-linux ontology-darwin

ontology-windows:
	GOOS=windows GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o ontology-windows-amd64.exe main.go

ontology-linux:
	GOOS=linux GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o ontology-linux-amd64 main.go

ontology-darwin:
	GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o ontology-darwin-amd64 main.go

tools-cross: tools-windows tools-linux tools-darwin

tools-windows: abi 
	GOOS=windows GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o sigsvr-windows-amd64.exe cmd-tools/sigsvr/sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr-windows-amd64.exe $(TOOLS)

tools-linux: abi 
	GOOS=linux GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o sigsvr-linux-amd64 cmd-tools/sigsvr/sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr-linux-amd64 $(TOOLS)

tools-darwin: abi 
	GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o sigsvr-darwin-amd64 cmd-tools/sigsvr/sigsvr.go
	@if [ ! -d $(TOOLS) ];then mkdir -p $(TOOLS) ;fi
	@mv sigsvr-darwin-amd64 $(TOOLS)

all-cross: ontology-cross tools-cross abi

format:
	$(GOFMT) -w main.go


docker: Makefile
	@echo "Building ontology docker"
	@$(DBUILD) --no-cache -t $(DOCKER_NS)/ontology:$(DOCKER_TAG) - < docker/Dockerfile 
	@touch $@

clean:
	rm -rf *.8 *.o *.out *.6 *exe coverage
	rm -rf ontology ontology-* tools

