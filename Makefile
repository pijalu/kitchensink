GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet
GOTEST=$(GOCMD) test
GOGEN=$(GOCMD) generate
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
EXE=./kitchensink

all: build test vet doc


${GOPATH}/bin/dep:
	$(GOGET) github.com/golang/dep/cmd/dep

${GOPATH}/bin/mockgen:
	$(GOGET) github.com/golang/mock/mockgen

generate: ${GOPATH}/bin/mockgen
	$(GOGEN) ./...

deps: generate godep

godep: ${GOPATH}/bin/dep
	dep ensure

test:
	$(GOCMD) test -v ./...

vet:
	$(GOVET) -v

$(EXE): deps
	$(GOBUILD) -ldflags="-s -w" -v -o $(EXE)

build: $(EXE)

doc: build
	$(EXE) doc ./documentation

clean:
	$(GOCLEAN)
	rm -rf mocks/*
	rm -rf documentation/*

run: $(EXE) 
	$(EXE) $(ARG)

install: all
	$(GOINSTALL)
