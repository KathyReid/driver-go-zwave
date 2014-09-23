GO_OPENZWAVE=github.com/ninjasphere/go-openzwave

all: build

# a rule that does the build in place here with whatever happens to be adjacent
here: deps fmt
	go install

fmt:
	gofmt -s -w *.go

clean:	clean-src
	go clean -i
	rm -rf bin/* || true
	rm -rf .gopath || true

clean-src:
	find . -name '*~' -exec rm {} \;

# does a clean build of all dependencies from git hub
build:
	scripts/build.sh

#
# A raw go build will not build the required dependency, so
# we use make to achieve this.
#
deps:
	go get -d $(GO_OPENZWAVE)
	cd $(GOPATH)/src/$(GO_OPENZWAVE) && make deps
	mkdir -p ninjapack/root/usr/lib
	cp $(GOPATH)/src/$(GO_OPENZWAVE)/openzwave/libopenzwave.so.1.0 ninjapack/root/usr/lib 
	mkdir -p ninjapack/root/usr/local/etc/openzwave/
	rsync --delete -ra $(GOPATH)/src/$(GO_OPENZWAVE)/openzwave/config/ ninjapack/root/usr/local/etc/openzwave/

test: install
	go test -v ./...

vet: install
	go vet ./...

.PHONY: all	dist clean test
