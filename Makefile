all: install

clean:
	rm -f bin/* || true
	rm -rf .gopath || true

install: deps
	go install

#
# A raw go build will not build the required dependency, so
# we use make to achieve this.
#
deps:
	cd ../go-openzwave && make	
	cp ../go-openzwave/openzwave/libopenzwave.so.1.0 ninjapack/root/usr/lib

test: install
	go test -v ./...

vet: install
	go vet ./...

.PHONY: all	dist clean test
