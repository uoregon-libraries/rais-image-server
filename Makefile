# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate getbuild binaries test format lint clean distclean docker plugins

# Default target builds binaries
all: binaries

# Generated code
generate: src/transform/rotation.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	gofmt -l -w -s src/transform/rotation.go

getbuild:
	go generate rais/src/version

# Binary building rules
binaries: src/transform/rotation.go getbuild
	go build -ldflags="-s -w" -o ./bin/rais-server rais/src/cmd/rais-server
	go build -ldflags="-s -w" -o ./bin/jp2info rais/src/cmd/jp2info

# Testing
test: getbuild
	go test rais/src/...

bench: getbuild
	go test -bench=. -benchtime=5s -count=2 rais/src/openjpeg rais/src/cmd/rais-server

format: getbuild
	find src/ -name "*.go" | xargs gofmt -l -w -s

lint: getbuild
	golint src/...
	go vet rais/src/...

# Cleanup
clean:
	rm -rf bin/
	rm -f src/transform/rotation.go
	rm -f src/version/build.go

# Generate the docker build container
docker: getbuild
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:build $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:latest-indev $(MakefileDir)

plugins: getbuild
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/s3-images.so rais/src/plugins/s3-images
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/external-images.so rais/src/plugins/external-images
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/datadog.so rais/src/plugins/datadog
	go build -ldflags="-s -w" -buildmode=plugin -o bin/plugins/json-tracer.so rais/src/plugins/json-tracer
