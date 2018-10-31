s3demo for RAIS
===

Setup
---

Run a simple, ugly "exhibit" of an entire S3 bucket!

1. Grab the RAIS codebase and go to this directory:

```bash
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server/docker/s3demo
```

2. Set up your environment for the docker stack

You can copy env-example to .env, and modify the required values **or** export
the necessary environment variables:

```bash
export AWS_ACCESS_KEY_ID="<your aws access key id>"
export AWS_SECRET_ACCESS_KEY="<your aws secret access key>"
export RAIS_S3ZONE="<AWS region / availability zone>"
export RAIS_S3BUCKET="<AWS S3 Bucket>"
export RAIS_IIIFURL="http://localhost/iiif"
```

3. Start the stack (`docker-compose up`) and visit `http://localhost`.  Gaze upon
your glorious images, lovingly served up by RAIS.

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
