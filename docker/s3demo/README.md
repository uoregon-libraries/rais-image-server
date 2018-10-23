s3demo for RAIS
===

Run a simple, ugly "exhibit" of an entire S3 bucket!

Open two terminal windows.  In terminal 1:

```bash
docker pull uolibraries/rais
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server/docker/s3demo
export AWS_ACCESS_KEY_ID="<your aws access key id>"
export AWS_SECRET_ACCESS_KEY="<your aws secret access key>"
export RAIS_S3ZONE="<AWS region / availability zone>"
export RAIS_S3BUCKET="<AWS S3 Bucket>"
export URL=http://localhost:12415/iiif
go build
./s3demo
```

In terminal 2, make sure to be in the same directory as above
(`rais-image-server/docker/s3demo`) and run compose:

```bash
docker-compose up
```

Visit `http://localhost:8080` and gaze upon your glorious images, lovingly
served up by RAIS.

This is a pretty weak demo, so be advised it's really just for testing, not
production use.  Some caveats:

- You won't be able to access this from anywhere but your local machine
- All caching is disabled in order to get realistic first-hit costs
- The images pulled from S3 live in ephemeral storage and will be deleted after
  you delete the RAIS container, again, to make it simple to get at realistic
  first-hit costs
- If you have non-images in your S3 bucket, behavior is undefined
- If you're running anything else on your server at port 8080 or 12415, this
  demo won't work as-is.  You may have to customize your setup (e.g., with a
  `docker-compose.override.yml` file)
- You must expose your AWS secrets to the environment for this to work with
  both the app and the dockerized RAIS container.  If this makes you
  uncomfortable, you can always dig into the setup manually.
- You must run `s3demo` in order to generate the `docker-compose.yml` file with
  the proper environment variables.  Running compose without first running the
  demo will result in either not having a compose file or using an old version
  of the compose file.
- The compose file is completely replaced each time `s3demo` is run.  Don't
  modify that file.  Use `docker-compose.override.yml` if you want to make the
  demo do anything "clever".
