GROUP := github.com/TheoBrigitte
NAME := ovh-dyndns

PKG := ${GROUP}/${NAME}

BIN := ${NAME}

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

VERSION := $(shell git describe --always --long --dirty || date)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: bin 

bin:
	GOOS=${OSTYPE} CGO_ENABLED=0 go build -i -v -o ${BIN} -ldflags="-s -w -X main.Version=${VERSION}" ${PKG}

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

.PHONY: run operator vet lint build
