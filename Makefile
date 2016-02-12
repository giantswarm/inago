PROJECT=formica

BUILD_PATH := $(shell pwd)/.gobuild
GS_PATH := $(BUILD_PATH)/src/github.com/giantswarm
GOPATH := $(BUILD_PATH)

BIN := $(PROJECT)

VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: all clean test ci-test

SOURCE=$(shell find . -name '*.go')

BUILD_COMMAND=go build -a -o $(BIN)
TEST_COMMAND=go test ./... -cover

all: $(BIN)

clean:
	rm -rf $(BUILD_PATH) $(BIN)

.gobuild:
	@mkdir -p $(GS_PATH)
	@rm -f $(GS_PATH)/$(PROJECT) && cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	#
	# Fetch private packages first (so `go get` skips them later)
	# @GOPATH=$(GOPATH) builder go get github.com/spf13/cobra
	# Pin versions of certain libs
	@builder get dep -b v0.10.2 git@github.com:coreos/fleet.git $(GOPATH)/src/github.com/coreos/fleet
	#
	@GOPATH=$(GOPATH) builder go get github.com/juju/errgo
	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v github.com/giantswarm/$(PROJECT)

$(BIN): $(SOURCE) VERSION .gobuild
	@echo Building inside Docker container for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:1.5 \
	    $(BUILD_COMMAND)

test:
	echo Testing inside Docker container for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:1.5 \
	    $(TEST_COMMAND)

lint:
	go vet -x

ci-build: $(SOURCE) VERSION .gobuild
	echo Building for $(GOOS)/$(GOARCH)
	$(BUILD_COMMAND)

ci-test:
	echo Testing for $(GOOS)/$(GOARCH)
	$(TEST_COMMAND)
