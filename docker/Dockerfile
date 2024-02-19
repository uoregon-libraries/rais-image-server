# This generates a production image for RAIS with the S3 plugin enabled
#
# Examples:
#
#     # Simple case: just build the latest production image
#     docker build --rm -t uolibraries/rais:latest -f ./docker/Dockerfile .
#
#     # Generate the build image to simplify local development
#     docker build --rm -t uolibraries/rais:build --target build -f ./docker/Dockerfile .
FROM golang:1 AS build
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Install all the build dependencies
RUN apt-get update -y && apt-get upgrade -y && \
    apt-get install -y libopenjp2-7-dev libmagickcore-dev git gcc make tar findutils

# Make sure the build box can lint code
RUN go install github.com/mgechev/revive@latest

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
ADD ./scripts /opt/rais-src/scripts
RUN make

# Manually build the ImageMagick plugin since we exclude it by default
RUN make bin/plugins/imagick-decoder.so



# Production image just installs runtime deps and copies in the binaries
FROM debian:buster AS production
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Stolen from mariadb dockerfile: add our user and group first to make sure
# their IDs get assigned consistently
RUN groupadd -r rais && useradd -r -g rais rais

# Install the core dependencies needed for both build and production
RUN apt-get update -y && apt-get upgrade -y && \
    apt-get install -y libopenjp2-7 imagemagick

ENV RAIS_TILEPATH /var/local/images
ENV RAIS_PLUGINS "*.so"
RUN touch /etc/rais.toml && chown rais:rais /etc/rais.toml
COPY --from=build /opt/rais-src/bin /opt/rais/

USER rais
EXPOSE 12415
ENTRYPOINT ["/opt/rais/rais-server"]
