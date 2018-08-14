# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate binaries test format lint clean distclean docker

# Default target builds binaries
all: binaries

# Dependencies all come via a single command
deps: vendor/src
vendor/src:
	gb vendor restore
	# Remove patented code from vendored area just in case
	rm vendor/src/github.com/hashicorp/golang-lru/arc*

# Generated code
generate: src/transform/rotation.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	gofmt -l -w -s src/transform/rotation.go

# Binary building rules
binaries: deps bin/rais-server bin/jp2info

# Build the server
bin/rais-server: src/transform/rotation.go src/* src/cmd/rais-server/*
	gb build cmd/rais-server

bin/jp2info: src/jp2info/* src/cmd/jp2info/*
	gb build cmd/jp2info

# Testing
test: deps
	gb test

bench: deps
	gb test -bench=. -run=XXX -v -test.benchtime=5s -test.count=2

format:
	find src/ -name "*.go" | xargs gofmt -l -w -s

lint:
	golint src/...

# Cleanup
clean:
	rm -f bin/*
	rm -rf pkg/
	rm -f src/transform/rotation.go

# (Re)build the separated docker containers
docker:
	docker build --rm -t rais-build:f28 -f docker/Dockerfile.build $(MakefileDir)
	docker run --rm -v $(MakefileDir):/opt/rais-src rais-build:f28 make clean
	docker run --rm -v $(MakefileDir):/opt/rais-src rais-build:f28 make
	docker build --rm -t uolibraries/rais:f28 -f docker/Dockerfile.prod $(MakefileDir)
