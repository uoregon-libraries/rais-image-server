s3demo for RAIS
===

Setup
---

Run a simple, ugly "exhibit" of an entire S3 bucket!

### Get RAIS

Grab the RAIS codebase and go to this directory:

```bash
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server/docker/s3demo
```

### Set up an S3 environment

We can do this the easy way or the hard way....

#### The easy way: minio

The demo is now set up to include "minio", an S3-compatible storage backend, by
default.  No actual S3 environment necessary!

Run the minio container:

`docker-compose up minio`

Create images:

- Browse to `http://localhost:9000`
- Log in with the acces key "awss3key" and the secret key "awsappsecret"
- Create a new bucket with the name "rais"
- Upload JP2s into this bucket

You can also use other s3 tools if the web interface for minio isn't to your
liking - you'll just have to specify the S3 endpoint as
`http://localhost:9000`.

When you're done, you can stop the minio container - it'll restart in the next
step anyway.

#### The hard way

You'll have to override the environment variables in `.env`.  The easiest way
is simply to copy `env-example` to `.env` and read what's there, customizing
AWS-specific data as necessary.

You'll also need to make sure you upload JP2 files into the bucket you
designated in your `.env` file.

A complete explanation of setting up and using AWS services is out of scope
here, however, so if you are unfamiliar with AWS, go with the easy way above.

### Start the stack

Run `docker-compose up` and visit `http://localhost`.  Gaze upon your glorious
images, lovingly served up by RAIS.

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
  `docker-compose.override.yml` file)

Development
---

Don't hack up the demo unless you want pain.  The demo server is a mess, and
the setup is a little hacky.  It's here to provide a quick demo, not showcase
elegant solutions to a problem.

If you are a masochist, however, make sure you re-run "docker-compose build"
anytime you change the codebase or the go templates.
