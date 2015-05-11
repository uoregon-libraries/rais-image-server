GO_NAMESPACE_DIR=$(GOPATH)/src/github.com/uoregon-libraries
GO_PROJECT_SYMLINK=$(GO_NAMESPACE_DIR)/newspaper-jp2-viewer
SYMLINK_EXISTS=$(GO_PROJECT_SYMLINK)/Makefile
GO_PROJECT_NAME=github.com/uoregon-libraries/newspaper-jp2-viewer
GOBIN=$(GOROOT)/bin/go

# Dependencies
IMGRESIZEDEP=github.com/nfnt/resize
IMGRESIZE=$(GOPATH)/src/$(IMGRESIZEDEP)

# All library files contribute to the need to recompile (except tests!  How do
# we skip those?)
SRCS := openjpeg/*.go iiif/*.go

.PHONY: all generate binaries test format lint clean distclean

# Default target builds binaries
all: binaries

# Generated code
generate: transform/rotation.go

transform/rotation.go: transform/generator.go transform/template.txt
	$(GOBIN) run transform/generator.go
	gofmt -l -w -s transform/rotation.go

# Dependency-getters
deps: $(IMGRESIZE)
$(IMGRESIZE):
	$(GOBIN) get $(IMGRESIZEDEP)

# dir/symlink creation - mandatory for any binary building to work
#
# We use symlink/main.go to avoid the symlink being a dependency - *any* change
# to directory listing will cause make to think it needs a rebuild if the rule
# is just the symlink itself
$(SYMLINK_EXISTS):
	mkdir -p $(GO_NAMESPACE_DIR)
	ln -s $(CURDIR) $(GO_PROJECT_SYMLINK)

# Binary building rules
binaries: bin/jp2tileserver
bin/jp2tileserver: $(SYMLINK_EXISTS) $(IMGRESIZE) $(SRCS) cmd/jp2tileserver/*.go transform/rotation.go
	$(GOBIN) build -o bin/jp2tileserver ./cmd/jp2tileserver

# Testing
test: $(SYMLINK_EXISTS) $(IMGRESIZE)
	$(GOBIN) test ./openjpeg ./cmd/jp2tileserver ./iiif ./fakehttp

format:
	find . -name "*.go" | xargs gofmt -l -w -s

lint:
	golint ./...

# Cleanup
clean:
	rm -f bin/*

distclean: clean
	rm -f $(GO_PROJECT_SYMLINK)
	rmdir --ignore-fail-on-non-empty $(GO_NAMESPACE_DIR)
