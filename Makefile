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

.PHONY: all clean test ci-test deps bin-dist install

SOURCE=$(shell find . -name '*.go')
INT_TESTS=$(shell find $(INT_TESTS_PATH) -name '*.t')

BUILD_COMMAND=go build\
 				-a -ldflags \
				"-X github.com/giantswarm/inago/cli.projectVersion=$(VERSION) -X github.com/giantswarm/inago/cli.projectBuild=$(COMMIT)" \
				-o $(BIN)
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
	@builder get dep -b 56b76bdf51f7708750eac80fa38b952bb9f32639 https://github.com/mattn/go-isatty.git $(GOPATH)/src/github.com/mattn/go-isatty
	@builder get dep -b e7da8edaa52631091740908acaf2c2d4c9b3ce90 https://github.com/golang/net.git $(GOPATH)/src/golang.org/x/net
	@builder get dep -b d2e44aa77b7195c0ef782189985dd8550e22e4de https://github.com/op/go-logging.git $(GOPATH)/src/github.com/op/go-logging

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

# With the dash before docker we don't exit if the 'docker run' returns with
# an error and run the rest of the target definition. Why? We want to destroy
# the test machine in any case.
int-test: $(BIN) $(INT_TESTS)
	@echo Running integration tests
	@echo Creating CoreOS integration test machine user-data
	cp $(VAGRANT_PATH)/user-data.sample $(VAGRANT_PATH)/user-data
	@echo Starting CoreOS integration test machine
	cd $(VAGRANT_PATH) && vagrant up
	sleep 10
	
	@echo Starting ssh-agent container
	-docker run \
	-d \
	--name=ssh-agent \
	whilp/ssh-agent:latest
	@echo Adding vagrant ssh key to ssh-agent container
	-docker run \
	--rm --volumes-from=ssh-agent \
	-v ~/.vagrant.d:/ssh \
	whilp/ssh-agent:latest \
	ssh-add /ssh/insecure_private_key
	
	-FLEET_ENDPOINT=$(FLEET_ENDPOINT) make internal-int-test
	
	@echo Destroying the ssh-agent container
	docker rm -f ssh-agent
	@echo Destroying the integration test machine
	cd $(VAGRANT_PATH) && vagrant destroy -f
	@echo Removing test machine user-data
	rm $(VAGRANT_PATH)/user-data

internal-int-test: $(BIN) $(INT_TESTS)
	docker run \
		--rm \
		-e FLEET_ENDPOINT=$(FLEET_ENDPOINT) \
		-e INAGO_TUNNEL_ENDPOINT=$(INAGO_TUNNEL_ENDPOINT) \
		-e SSH_AUTH_SOCK=/root/.ssh/socket \
		--volumes-from ssh-agent  \
		-v $(CURDIR)/$(BIN):/usr/local/bin/$(BIN) \
		-v $(INT_TESTS_PATH):$(INT_TESTS_PATH) \
		zeisss/cram-docker \
		-v $(INT_TESTS_PATH)

bin-dist: $(SOURCE) VERSION .gobuild
	# Remove any old bin-dist or build directories
	rm -rf bin-dist build
	
	# Build for all supported OSs
	for OS in darwin linux; do \
		rm -f $(BIN); \
		GOOS=$$OS make $(BIN); \
		mkdir -p build/$$OS bin-dist; \
		cp README.md build/$$OS/; \
		cp LICENSE build/$$OS/; \
		cp $(BIN) build/$$OS/; \
		tar czf bin-dist/$(BIN).$(VERSION).$$OS.tar.gz -C build/$$OS .; \
	done

install: $(BIN)
	cp $(BIN) /usr/local/bin/
