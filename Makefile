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
SRCS := openjpeg/*.go

.PHONY: all binaries test clean distclean

# Default target builds binaries
all: binaries

# Dependency-getters
deps: $(GOPATH)/src/github.com/nfnt/resize
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
binaries: bin/jp2tileserver bin/verifyJP2s
bin/jp2tileserver: $(SYMLINK_EXISTS) $(IMGRESIZE) $(SRCS) cmd/jp2tileserver/*.go
	$(GOBIN) build -o bin/jp2tileserver ./cmd/jp2tileserver
bin/verifyJP2s: $(SYMLINK_EXISTS) $(IMGRESIZE) $(SRCS) cmd/verifyJP2s/*.go
	$(GOBIN) build -o bin/verifyJP2s ./cmd/verifyJP2s

# Testing
test: $(SYMLINK_EXISTS) $(IMGRESIZE)
	$(GOBIN) test ./openjpeg

# Cleanup
clean:
	rm -f bin/*

distclean: clean
	rm -f $(GO_PROJECT_SYMLINK)
	rmdir --ignore-fail-on-non-empty $(GO_NAMESPACE_DIR)
