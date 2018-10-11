Docker
===

This directory contains everything necessary to run RAIS under Docker,
including a test image for the docker-compose-based demo.

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
