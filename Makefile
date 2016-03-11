PROJECT=inago

BUILD_PATH := $(shell pwd)/.gobuild
GS_PATH := $(BUILD_PATH)/src/github.com/giantswarm
GOPATH := $(BUILD_PATH)
INT_TESTS_PATH := $(shell pwd)/int-tests
VAGRANT_PATH := $(INT_TESTS_PATH)/vagrant

GOVERSION=1.6

BIN := $(PROJECT)ctl

VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: all clean test ci-test deps

SOURCE=$(shell find . -name '*.go')
INT_TESTS=$(shell find $(INT_TESTS_PATH) -name '*.t')

BUILD_COMMAND=go build -a -o $(BIN)
TEST_COMMAND=./go.test.sh

all: $(BIN)

clean:
	rm -rf $(BUILD_PATH) $(BIN)

.gobuild:
	@mkdir -p $(GS_PATH)
	@rm -f $(GS_PATH)/$(PROJECT) && cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	#
	# Fetch and pin packages
	@builder get dep -b 76516ab4ae194e37aaae9c1f2fa5090553e541f3 https://github.com/coreos/fleet.git $(GOPATH)/src/github.com/coreos/fleet
	@builder get dep -b 08cceb5d0b5331634b9826762a8fd53b29b86ad8 https://github.com/juju/errgo.git $(GOPATH)/src/github.com/juju/errgo
	@builder get dep -b 65a708cee0a4424f4e353d031ce440643e312f92 https://github.com/spf13/cobra.git $(GOPATH)/src/github.com/spf13/cobra
	@builder get dep -b 7f60f83a2c81bc3c3c0d5297f61ddfa68da9d3b7 https://github.com/spf13/pflag.git $(GOPATH)/src/github.com/spf13/pflag
	@builder get dep -b 983d3a5fab1bf04d1b412465d2d9f8430e2e917e https://github.com/ryanuber/columnize.git $(GOPATH)/src/github.com/ryanuber/columnize
	@builder get dep -b e673fdd4dea8a7334adbbe7f57b7e4b00bdc5502 https://github.com/satori/go.uuid.git $(GOPATH)/src/github.com/satori/go.uuid

	@builder get dep https://github.com/onsi/gomega.git $(GOPATH)/src/github.com/onsi/gomega
	@builder get dep https://github.com/stretchr/testify.git $(GOPATH)/src/github.com/stretchr/testify
	@builder get dep https://github.com/davecgh/go-spew.git $(GOPATH)/src/github.com/davecgh/go-spew
	@builder get dep https://github.com/pmezard/go-difflib.git $(GOPATH)/src/github.com/pmezard/go-difflib
	@builder get dep https://github.com/stretchr/objx.git $(GOPATH)/src/github.com/stretchr/objx
deps:
	@${MAKE} -B -s .gobuild

$(BIN): $(SOURCE) VERSION .gobuild
	@echo Building inside Docker container for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -ti \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:$(GOVERSION) \
	    $(BUILD_COMMAND)

test: $(SOURCE) VERSION .gobuild
	@echo Testing inside Docker container for $(GOOS)/$(GOARCH)
	docker run \
	    --rm \
	    -ti \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code \
	    golang:$(GOVERSION) \
	    $(TEST_COMMAND)

lint:
	go vet -x ./...
	golint ./...

ci-build: $(SOURCE) VERSION .gobuild
	echo Building for $(GOOS)/$(GOARCH)
	$(BUILD_COMMAND)

ci-test: $(SOURCE) VERSION .gobuild
	echo Testing for $(GOOS)/$(GOARCH)
	$(TEST_COMMAND)
	
# Use with `GOOS=linux FLEET_ENDPOINT=http://192.168.99.1:49153/ make int-test`
# Set fleet endpoint to a fleet API endpoint available to the container.
int-test: $(BIN) $(INT_TESTS)
	@echo Running integration tests
	@echo Creating CoreOS integration test machine user-data
	cp $(VAGRANT_PATH)/user-data.sample $(VAGRANT_PATH)/user-data
	@echo Starting CoreOS integration test machine
	cd $(VAGRANT_PATH) && vagrant up
	sleep 10
	# With the dash before docker we don't exit if the 'docker run' returns with
	# an error and run the rest of the target definition. Why? We want to destroy
	# the test machine in any case.
	-docker run \
		--rm \
		-ti \
		-e FLEET_ENDPOINT=$(FLEET_ENDPOINT) \
		-v $(CURDIR)/$(BIN):/usr/local/bin/$(BIN) \
		-v $(INT_TESTS_PATH):$(INT_TESTS_PATH) \
		zeisss/cram-docker \
		-v $(INT_TESTS_PATH)
	@echo Destroying the integration test machine
	cd $(VAGRANT_PATH) && vagrant destroy -f
	@echo Removing test machine user-data
	rm $(VAGRANT_PATH)/user-data
