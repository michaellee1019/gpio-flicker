BIN_OUTPUT_PATH = bin
TOOL_BIN = bin/gotools/$(shell uname -s)-$(shell uname -m)
UNAME_S ?= $(shell uname -s)
GOPATH = $(HOME)/go/bin
export PATH := ${PATH}:$(GOPATH)

build: format update-rdk
	rm -f $(BIN_OUTPUT_PATH)/gpio-flicker
	GOOS=$(VIAM_BUILD_OS) GOARCH=$(VIAM_BUILD_ARCH) go build -o $(BIN_OUTPUT_PATH)/gpio-flicker main.go

module.tar.gz: build
	rm -f $(BIN_OUTPUT_PATH)/module.tar.gz
	tar czf $(BIN_OUTPUT_PATH)/module.tar.gz $(BIN_OUTPUT_PATH)/gpio-flicker meta.json

reload-build:
	mkdir -p $(BIN_OUTPUT_PATH)
	rm -f $(BIN_OUTPUT_PATH)/module.tar.gz
	tar -czf $(BIN_OUTPUT_PATH)/module.tar.gz main.go meta.json reload.sh Makefile models/* go.mod go.sum

reload-setup:
	test -f go1.23.5.linux-arm64.tar.gz || wget https://go.dev/dl/go1.23.5.linux-arm64.tar.gz
	sudo tar -C /usr/local -xzf go1.23.5.linux-arm64.tar.gz
	sudo apt-get install -y apt-utils coreutils tar libnlopt-dev libjpeg-dev pkg-config

setup:
	if [ ! -f .installed ]; then \
		if [ "$(UNAME_S)" = "Linux" ]; then \
			sudo apt-get update; \
			sudo apt-get install -y golang apt-utils coreutils tar libnlopt-dev libjpeg-dev pkg-config; \
		fi; \
		# remove unused imports \
		go install golang.org/x/tools/cmd/goimports@latest; \
		find . -name '*.go' -exec $(GOPATH)/goimports -w {} +; \
		touch .installed; \
	else \
		echo "Already installed. Delete .installed file to force reinstallation."; \
	fi

clean:
	rm -rf $(BIN_OUTPUT_PATH)/gpio-flicker $(BIN_OUTPUT_PATH)/module.tar.gz gpio-flicker

format:
	gofmt -w -s .

update-rdk:
	go get go.viam.com/rdk@latest
	go mod tidy
