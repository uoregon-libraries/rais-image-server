# s3demo for RAIS

This folder is dedicated to a mini-asset-manager-like application that presents
a local file alongside any number of cloud files in an S3-compatible object
store.

## Setup

Run a simple, ugly "exhibit" of an entire S3 bucket!

### Get RAIS

Grab the RAIS codebase and go to this directory:

```bash
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server/docker/s3demo
```

Copy and edit `compose.override.example.yml` to work for your environment.

### Set up an minio (a local S3-compatible) environment

The demo is set up to include "minio", an S3-compatible storage backend, by
default. No actual S3 environment necessary!

Run the minio container:

`docker compose up minio`

Create images:

- Browse to `http://localhost:9000` to make sure the service is working
- If you are able to use the Web UI (it's a wonderful combination of super ugly
  and completely inaccessible):
  - Log in as "admin" with password "admin123"
  - Create a new bucket with the name "rais"
  - Upload JP2s into this bucket

You can also use other s3 tools if the web interface for minio isn't usable for
you. You'll just have to specify the S3 endpoint as `http://localhost:9000`.

The command-line tool `mc` can be used something like this (from your host, not
one of the containers!):

```bash
go install github.com/minio/mc@latest

# Disable mc's built-in pager: it's got fewer features than `less` and `more`,
# and somehow makes terminal accessibility worse than those tools
export MC_DISABLE_PAGER=1

mc alias set rais http://localhost:9000 admin admin123
mc admin accesskey create rais --access-key access-key --secret-key secret-key
mc mb rais/rais
mc cp --recursive ../images/jp2tests/ rais/rais/

# Verify the images are there
mc ls rais/rais
```

### Or set up AWS S3

If you want a real test, you're mostly on your own.

You'll have to override the environment variables in `.env`. The easiest way
is simply to copy `env-example` to `.env` and read what's there, customizing
AWS-specific data as necessary.

You'll also need to make sure you upload JP2 files into the bucket you
designated in your `.env` file.

A complete explanation of setting up and using AWS services is out of scope
here, however, so if you are unfamiliar with AWS, go with the easy way above.

### Start the stack

Run `docker compose up -d minio && sleep 1 && docker compose up -d` and visit
`http://localhost`. Gaze upon your glorious images, lovingly served up by
RAIS.

You *must* make sure your minio container is running first (hence the weird
command above), as the s3demo will crash on startup if minio isn't ready.
