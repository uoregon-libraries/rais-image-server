s3demo for RAIS
===

This folder is dedicated to a mini-asset-manager-like application that presents
a local file alongside any number of cloud files in an S3-compatible object
store.

Setup
---

Run a simple, ugly "exhibit" of an entire S3 bucket!

### Get RAIS

Grab the RAIS codebase and go to this directory:

```bash
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server/docker/s3demo
```

### Optional: build a docker image

You may want to build your own docker image rather than using the latest stable
version from dockerhub, especially if testing out plugins or other code that
requires a recompile.  Building an image typically means running `make docker`
and/or `make docker-alpine` *from the RAIS root directory*.  Note that the
alpine image is *much* faster to build, but doesn't contain any plugins.

Once that's done, you will want to put an override for the s3 demo so it uses
your local image.  Something like this can be put into
`compose.override.yml` in this (the s3demo) directory:

```
networks:
  internal:
  external:

services:
  rais:
    image: uolibraries/rais:latest-alpine
```

### Set up an S3 environment

We can do this the easy way or the hard way....

#### The easy way: minio

The demo is now set up to include "minio", an S3-compatible storage backend, by
default.  No actual S3 environment necessary!

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

The command-line tool `mc` can be used something like this:

```bash
go install github.com/minio/mc@latest

# Disable mc's built-in pager: it's got fewer features than `less` and `more`,
# and somehow makes terminal accessibility worse than those tools
export MC_DISABLE_PAGER=1

mc alias set rais http://localhost:9000 admin admin123
mc admin accesskey create rais --access-key access-key --secret-key secret-key
mc mb rais/rais
mc cp ../images/jp2tests/* rais/rais/
```

#### The hard way

You'll have to override the environment variables in `.env`.  The easiest way
is simply to copy `env-example` to `.env` and read what's there, customizing
AWS-specific data as necessary.

You'll also need to make sure you upload JP2 files into the bucket you
designated in your `.env` file.

A complete explanation of setting up and using AWS services is out of scope
here, however, so if you are unfamiliar with AWS, go with the easy way above.

### Start the stack

Run `docker compose up -d minio && sleep 1 && docker compose up -d` and visit
`http://localhost`.  Gaze upon your glorious images, lovingly served up by
RAIS.

You *must* make sure your minio container is running first (hence the weird
command above), as the s3demo will crash on startup if minio isn't ready.

Caveats
---

This is a pretty weak demo, so be advised it's really just for testing, not
production use.  Some caveats:

- The images pulled from S3 live in ephemeral storage and will be deleted after
  you delete the RAIS container.  This makes it simple to get at realistic
  first-hit costs
- If you have non-images in your S3 bucket, behavior is undefined
- If you're running anything else on your server at port 80, this demo won't
  work as-is.  You may have to customize your setup (e.g., with a
  `compose.override.yml` file)

Development
---

Don't hack up the demo unless you want pain.  The demo server is a mess, and
the setup is a little hacky.  It's here to provide a quick demo, not showcase
elegant solutions to a problem.

If you are a masochist, however, make sure you re-run "docker compose build"
anytime you change the codebase or the go templates.
