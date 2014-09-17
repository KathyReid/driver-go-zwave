all: build

# a rule that does the build in place here with whatever happens to be adjacent
here: deps
	go install

clean:
	go clean -i
	rm -rf bin || true
	rm -rf .gopath || true

# does a clean build of all dependencies from git hub
build:
	scripts/build.sh

#
# A raw go build will not build the required dependency, so
# we use make to achieve this.
#
deps:
	go get github.com/ninjasphere/go-openzwave 
	cd ../go-openzwave && make deps
	mkdir -p ninjapack/root/usr/lib
	cp ../go-openzwave/openzwave/libopenzwave.so.1.0 ninjapack/root/usr/lib 
	mkdir -p ninjapack/root/usr/local/etc/openzwave/
	rsync --delete -ra ../go-openzwave/openzwave/config/ ninjapack/root/usr/local/etc/openzwave/

test: install
	go test -v ./...

vet: install
	go vet ./...

.PHONY: all	dist clean test
