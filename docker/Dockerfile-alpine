# This generates a production alpine image for RAIS
#
# Example:
#
#     docker build --rm -t uolibraries/rais:latest-alpine -f ./docker/Dockerfile-alpine .
FROM golang:1-alpine AS build
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Install all the build dependencies
RUN apk add --no-cache openjpeg-dev
RUN apk add --no-cache imagemagick6-dev
RUN apk add --no-cache git
RUN apk add --no-cache gcc
RUN apk add --no-cache make
RUN apk add --no-cache tar
RUN apk add --no-cache curl

# Installing GraphicsMagick is wholly unnecessary, but helps when using the
# build box for things like converting images.  Since opj2_encode doesn't
# support all formats, and ImageMagick has been iffy with some conversions for
# us, "gm convert" is a handy fallback.  As this is a multi-stage dockerfile,
# this installation doesn't hurt the final production docker image.
RUN apk add --no-cache graphicsmagick

# Go comes after other installs to avoid re-pulling the more expensive
# dependencies when changing Go versions
RUN apk add --no-cache musl-dev

# Make sure the build box can lint code
RUN go get -u golang.org/x/lint/golint

# Add the go mod stuff first so we aren't re-downloading dependencies except
# when they actually change
WORKDIR /opt/rais-src
ADD ./go.mod /opt/rais-src/go.mod
ADD ./go.sum /opt/rais-src/go.sum
RUN go mod download

# Make sure we don't just add every little thing, otherwise unimportant changes
# trigger a rebuild
ADD ./Makefile /opt/rais-src/Makefile
ADD ./src /opt/rais-src/src
RUN make binaries plugins

# Production image just installs runtime deps and copies in the binaries
FROM alpine:3.8 AS production
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Add our user and group first to make sure their IDs get assigned consistently
RUN addgroup -S rais && adduser -S rais -G rais

# Deps
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN apk add --no-cache openjpeg imagemagick6

ENV RAIS_TILEPATH /var/local/images
RUN touch /etc/rais.toml && chown rais:rais /etc/rais.toml
COPY --from=build /opt/rais-src/bin /opt/rais/

#USER rais
EXPOSE 12415
ENTRYPOINT ["/opt/rais/rais-server"]
