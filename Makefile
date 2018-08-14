# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate binaries test format lint clean distclean docker

# Default target builds binaries
all: binaries

# Generated code
generate: src/transform/rotation.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	gofmt -l -w -s src/transform/rotation.go

# Binary building rules
binaries: bin/rais-server bin/jp2info

bin/rais-server: src/transform/rotation.go
	go build -o ./bin/rais-server rais/src/cmd/rais-server

bin/jp2info:
	go build -o ./bin/jp2info rais/src/cmd/jp2info

# Testing
test:
	go test rais/src/...

bench:
	go test -bench=. -run=XXX -v -test.benchtime=5s -test.count=2

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
	docker pull uolibraries/rais
	docker-compose build rais-build
	docker-compose run --rm rais-build make clean
	docker-compose run --rm rais-build make
	docker build --rm -t uolibraries/rais:f28 -f docker/Dockerfile.prod $(MakefileDir)
