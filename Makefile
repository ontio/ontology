GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
Minversion := $(shell date)
BUILD_PAR = -ldflags "-X main.Version=$(VERSION)" #-race

all:
	$(GC)  $(BUILD_PAR) -o node main.go
	$(GC)  $(BUILD_PAR) nodectl.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6
