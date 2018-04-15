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

${GOPATH}/bin/mockgen:
	$(GOGET) github.com/golang/mock/mockgen

generate: ${GOPATH}/bin/mockgen
	$(GOGEN) ./...

deps: generate
	$(GOGET) ./...

test:
	$(GOCMD) test -v ./...

vet:
	$(GOVET) -v

$(EXE): deps
	$(GOBUILD) -v -o $(EXE)

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



