GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
Minversion := $(shell date)
BUILD_NODE_PAR = -ldflags "-X DNA/common/config.Version=$(VERSION)" #-race
BUILD_NODECTL_PAR = -ldflags "-X main.Version=$(VERSION)"

all:
	$(GC)  $(BUILD_NODE_PAR) -o node main.go
	$(GC)  $(BUILD_NODECTL_PAR) nodectl.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6
