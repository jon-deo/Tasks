# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOLINT=golint
BINARY_NAME=student-report-service

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

run:
	$(GOBUILD) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	find . -type f -name '*.test' -delete

# Install dependencies
deps:
	$(GOGET) -v ./...

# Format code
fmt:
	$(GOFMT) ./...

# Vet code
vet:
	$(GOVET) ./...

# Run linter
lint:
	$(GOLINT) ./...

# Build Docker image
docker-build:
	docker build -t student-report-service .

# Run Docker container
docker-run:
	docker run -p 5008:5008 student-report-service

.PHONY: all build test run clean deps fmt vet lint docker-build docker-run