# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate binaries test format lint clean distclean docker plugins

# Default target builds binaries
all: binaries

# Generated code
generate: src/transform/rotation.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	gofmt -l -w -s src/transform/rotation.go

# Binary building rules
binaries: src/transform/rotation.go
	go build -ldflags="-s -w" -o ./bin/rais-server rais/src/cmd/rais-server
	go build -ldflags="-s -w" -o ./bin/jp2info rais/src/cmd/jp2info

# Testing
test:
	go test rais/src/...

bench:
	go test -bench=. -benchtime=5s -count=2 rais/src/openjpeg rais/src/cmd/rais-server

format:
	find src/ -name "*.go" | xargs gofmt -l -w -s

lint:
	golint src/...
	go vet rais/src/...

# Cleanup
clean:
	rm -rf bin/
	rm -f src/transform/rotation.go

# Generate the docker build container
docker:
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:build $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:latest-indev $(MakefileDir)

plugins:
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/s3-images.so rais/src/plugins/s3-images
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/external-images.so rais/src/plugins/external-images
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/datadog.so rais/src/plugins/datadog
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/json-tracer.so rais/src/plugins/json-tracer
