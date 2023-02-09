GROUP := github.com/TheoBrigitte
NAME := ovh-dyndns

PKG := ${GROUP}/${NAME}

BIN := ${NAME}

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

VERSION := $(shell git describe --always --long --dirty || date)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: bin 

build:
	GOOS=${OSTYPE} CGO_ENABLED=0 go build -v -o ${BIN} -ldflags="-s -w -X main.Version=${VERSION}" ${PKG}

install:
	GOOS=${OSTYPE} CGO_ENABLED=0 go install -v -ldflags="-s -w -X main.Version=${VERSION}" ${PKG}

test:
	@go test -short ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

clean:
	-@rm ${BIN}

.PHONY: build install test vet lint clean
