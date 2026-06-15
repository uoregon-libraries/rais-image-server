# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate binaries test format lint clean distclean docker

BUILD := $(shell git describe --tags)

# Default target builds binaries
all: cgo binaries

# Security check
.PHONY: audit
audit:
	go tool govulncheck $$(go list ./src/... | grep -v 'src/plugins/imagick')

.PHONY: cgo
cgo:
	./scripts/can_cgo.sh

# Generated code
generate: src/transform/rotation.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	go fmt src/transform/rotation.go

# Binary building rules
binaries: src/transform/rotation.go rais-server jp2info bin/plugins/json-tracer.so

rais-server:
	go build -ldflags="-s -w -X rais/src/version.Version=$(BUILD)" -o ./bin/rais-server rais/src/cmd/rais-server

jp2info:
	go build -ldflags="-s -w -X rais/src/version.Version=$(BUILD)" -o ./bin/jp2info rais/src/cmd/jp2info

# Testing; imagick plugins are excluded because they need ImageMagick dev
# libraries, which we don't want to require for routine test runs / CI
test:
	go test $$(go list ./src/... | grep -v 'src/plugins/imagick')

bench:
	go test -bench=. -benchmem -count=2 rais/src/openjpeg rais/src/cmd/rais-server

format:
	find src/ -name "*.go" | xargs go tool goimports -l -w

# Vet skips the imagick plugins for the same reason "test" does: vet has to
# compile cgo packages, and we don't want ImageMagick required everywhere
.PHONY: vet
vet:
	go vet $$(go list ./src/... | grep -v 'src/plugins/imagick')

lint: vet
	go tool revive src/...

# Cleanup
clean:
	rm -rf bin/
	rm -f src/transform/rotation.go

distclean: clean
	go clean -modcache -testcache -cache
	docker rmi uolibraries/rais:build || true
	docker rmi uolibraries/rais:build-alpine || true
	docker rmi uolibraries/rais:dev || true
	docker rmi uolibraries/rais:dev-alpine || true

# Generate the docker images
docker: | generate
	docker pull golang:1
	docker pull golang:1-alpine
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile -t rais:build $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:dev $(MakefileDir)
	make docker-alpine

# Build just the alpine image for cases where we want to get this updated / cranked out fast
docker-alpine: | generate
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile-alpine -t rais:build-alpine $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile-alpine -t uolibraries/rais:dev-alpine $(MakefileDir)

# Build plugins on any change to their directory or their go files
bin/plugins/%.so : src/plugins/% src/plugins/%/*.go
	go build -ldflags="-s -w" -buildmode=plugin -o $@ rais/$<
