PROJECT=infra-tmpl-go

BUILD_PATH := $(shell pwd)/.gobuild

GS_PATH := "$(BUILD_PATH)/src/github.com/giantswarm"

BIN=$(PROJECT)

.PHONY=clean run-test get-deps update-deps fmt run-tests

GOPATH := $(BUILD_PATH)
GOVERSION = 1.4.2-cross

SOURCE=$(shell find . -name '*.go')

all: get-deps $(BIN)

ci: clean all run-tests

clean:
	rm -rf $(BUILD_PATH) $(BIN)

get-deps: .gobuild

.gobuild:
	mkdir -p $(GS_PATH)
	cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	#
	# Fetch internal libraries
	#
	# Fetch public dependencies via `go get`
	@GOPATH=$(GOPATH) builder go get github.com/juju/errgo
	@GOPATH=$(GOPATH) builder go get github.com/spf13/viper
	@GOPATH=$(GOPATH) builder go get golang.org/x/crypto/openpgp
	@GOPATH=$(GOPATH) builder go get github.com/mitchellh/go-homedir
	@GOPATH=$(GOPATH) builder go get github.com/DisposaBoy/JsonConfigReader
	@GOPATH=$(GOPATH) builder go get github.com/spf13/pflag

$(BIN): $(SOURCE)
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -w /usr/code \
	    golang:$(GOVERSION) \
	    go build -o $(BIN)

run-tests:
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -w /usr/code \
	    golang:$(GOVERSION) \
	    go test ./...

run-test:
	if test "$(test)" = "" ; then \
		echo "missing test parameter, that is, path to test folder e.g. './middleware/v1/'."; \
		exit 1; \
	fi
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -w /usr/code \
	    golang:$(GOVERSION) \
	    go test -v $(test)

fmt:
	gofmt -l -w .
