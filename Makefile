GOFMT=gofmt
GC=go build

all:
	$(GC) main.go
format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6
