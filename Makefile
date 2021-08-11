# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=xstpl

all: build

build:
	$(GOBUILD) -o ./$(BINARY_NAME) -tags=jsoniter -v ./
