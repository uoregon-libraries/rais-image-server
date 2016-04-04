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
binaries: deps bin/rais-server

# Build the server.  Note that the "gb build openjpeg" is necessary to avoid an
# error when building the server before openjpeg has been compiled
bin/rais-server: src/transform/rotation.go src/*
	gb build openjpeg
	gb build rais-server

# Testing
test: deps
	gb test

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
	docker build --rm -t rais-build -f docker/Dockerfile.build $(MakefileDir)
	docker run --rm -v $(MakefileDir):/opt/rais-src rais-build
	docker build --rm -t uolibraries/rais:prod -f docker/Dockerfile.prod $(MakefileDir)
