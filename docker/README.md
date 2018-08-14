Docker
===

This directory contains split up docker files:

- `Dockerfile.build` relies on the fedora image and adds the build dependencies
  needed to compile RAIS
- `Dockerfile.prod` is a simpler image with only runtime dependencies, and
  assumes it can copy `rais-server` from a "bin" subdirectory.

It also contains demo files/folders:

- `demo-rais-entry.sh` and `demo-web-entry.sh` should be left alone, and simply
  bootstrap the docker setup for demo use.
- `ngingx.conf` is used for serving up static content and proxying to RAIS.
  This can be used as an example of how you might set up RAIS with ngingx in
  production.
- `images/` contains a single test image, but any JP2 files you add will be
  served up in the demo stack
- `static/` contains the Open Seadragon sources as well as static HTML for nginx

Building docker images
---

The easiest way to use these is from the parent directory's `Makefile` via
`make docker`.

Running the demo
---

From the project root:

```bash
# Set up your local server's URL if "localhost" won't suffice for any reason
export URL=http://192.168.0.5

# Copy images into images/
cp /some/jp2/sources/*.jp2 ./docker/images/

# Run nginx and RAIS
docker-compose up
```
