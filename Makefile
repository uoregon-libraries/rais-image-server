# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate force-getbuild binaries test format lint clean distclean docker plugins

# Default target builds binaries
all: binaries

# Generated code
generate: src/transform/rotation.go src/version/build.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	go fmt src/transform/rotation.go

force-getbuild:
	rm -f src/version/build.go
	make src/version/build.go

src/version/build.go:
	go generate rais/src/version

# Binary building rules
binaries: src/transform/rotation.go src/version/build.go plugins rais-server jp2info

rais-server:
	go build -ldflags="-s -w" -o ./bin/rais-server rais/src/cmd/rais-server

jp2info:
	go build -ldflags="-s -w" -o ./bin/jp2info rais/src/cmd/jp2info

# Testing
test: src/version/build.go
	go test rais/src/...

bench: src/version/build.go
	go test -bench=. -benchtime=5s -count=2 rais/src/openjpeg rais/src/cmd/rais-server

format: src/version/build.go
	find src/ -name "*.go" | xargs gofmt -l -w -s

lint: src/version/build.go
	revive src/...
	go vet rais/src/...

# Cleanup
clean:
	rm -rf bin/
	rm -f src/transform/rotation.go
	rm -f src/version/build.go

distclean: clean
	go clean -modcache -testcache -cache
	docker rmi uolibraries/rais:build || true
	docker rmi uolibraries/rais:build-alpine || true
	docker rmi uolibraries/rais:dev || true
	docker rmi uolibraries/rais:dev-alpine || true

# Generate the docker images
docker: | force-getbuild generate
	docker pull golang:1
	docker pull golang:1-alpine
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile -t rais:build $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:dev $(MakefileDir)
	make docker-alpine

# Build just the alpine image for cases where we want to get this updated / cranked out fast
docker-alpine: | force-getbuild generate
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile-alpine -t rais:build-alpine $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile-alpine -t uolibraries/rais:dev-alpine $(MakefileDir)

# Build plugins on any change to their directory or their go files
bin/plugins/%.so : src/plugins/% src/version/build.go src/plugins/%/*.go
	go build -ldflags="-s -w" -buildmode=plugin -o $@ rais/$<

# Build the plugins that don't have external dependencies
PLUGS := $(shell ./scripts/pluglist.sh)
plugins: $(PLUGS)
