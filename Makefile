IMAGE_NAME := github.com/lingdie/image-policy-webhook
IMAGE_TAG := latest

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

BINARY_NAME=bin/main

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) cmd/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) cmd/main.go
	./$(BINARY_NAME) --debug

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: all build clean run docker-build docker-push
