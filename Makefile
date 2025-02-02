BINARY_NAME := blog
BIN_DIR := /usr/local/bin
SRCS := $(shell git ls-files '*.go')
LDFLAGS := "-X github.com/babarot/blog/internal/cmd.version=$(shell git describe --tags --abbrev=0 --always) -X github.com/babarot/blog/internal/cmd.revision=$(shell git rev-parse --verify --short HEAD) -X github.com/babarot/blog/internal/cmd.buildDate=$(shell date "+%Y-%m-%d")"

all: build

test: $(SRCS)
	go test ./...

deps:
	go mod tidy

build: deps $(BINARY_NAME)

$(BINARY_NAME): $(SRCS)
	go build -ldflags $(LDFLAGS)

install:
	go install -ldflags $(LDFLAGS)

sys-install: build
	sudo install $(BINARY_NAME) /usr/local/bin

clean:
	rm -f $(BINARY_NAME)

.PHONY: all test deps build install clean
