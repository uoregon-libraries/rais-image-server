This directory contains split up docker files:

- `Dockerfile.build` relies on the fedora image and adds the build dependencies
  needed to compile RAIS
- `Dockerfile.prod` is a simpler image with only runtime dependencies, and
  assumes it can copy `rais-server` from a "bin" subdirectory.

The easiest way to use these is from the parent directory's `Makefile` via
`make docker`
