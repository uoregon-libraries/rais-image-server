# Makefile directory
MakefileDir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all generate getbuild binaries test format lint clean distclean docker plugins

# Default target builds binaries
all: binaries

# Generated code
generate: src/transform/rotation.go src/version/build.go

src/transform/rotation.go: src/transform/generator.go src/transform/template.txt
	go run src/transform/generator.go
	gofmt -l -w -s src/transform/rotation.go

getbuild: src/version/build.go

src/version/build.go:
	go generate rais/src/version
	@chmod a+w src/version/build.go 2>/dev/null || true

# Binary building rules
binaries: src/transform/rotation.go src/version/build.go plugins
	go build -ldflags="-s -w" -o ./bin/rais-server rais/src/cmd/rais-server
	go build -ldflags="-s -w" -o ./bin/jp2info rais/src/cmd/jp2info
	@chmod -R a+w bin/ 2>/dev/null || true

# Testing
test: src/version/build.go
	go test rais/src/...

bench: src/version/build.go
	go test -bench=. -benchtime=5s -count=2 rais/src/openjpeg rais/src/cmd/rais-server

format: src/version/build.go
	find src/ -name "*.go" | xargs gofmt -l -w -s

lint: src/version/build.go
	golint src/...
	go vet rais/src/...

# Cleanup
clean:
	rm -rf bin/
	rm -f src/transform/rotation.go
	rm -f src/version/build.go

# Generate the docker images
docker: | clean generate
	docker build --rm --target build -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:build $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile -t uolibraries/rais:latest-indev $(MakefileDir)
	docker build --rm -f $(MakefileDir)/docker/Dockerfile-alpine -t uolibraries/rais:latest-alpine $(MakefileDir)

# Build plugins on any change to their directory or their go files
bin/plugins/%.so : src/plugins/% src/version/build.go src/plugins/%/*.go
	go build -ldflags="-s -w" -buildmode=plugin -o $@ rais/$<

# Build all the plugins with horrible shell magic!
PLUGS := $(shell ./scripts/pluglist.sh)
plugins: $(PLUGS)
	@chmod -R a+w bin/ || true
