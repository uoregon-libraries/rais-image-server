# This generates a production image for RAIS with the S3 plugin enabled
#
# Examples:
#
#     # Simple case: just build the latest production image
#     docker build --rm -t uolibraries/rais:latest -f ./docker/Dockerfile .
#
#     # Generate the build image to simplify local development
#     docker build --rm -t uolibraries/rais:build --target build -f ./docker/Dockerfile .
FROM fedora:32 AS build
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Install all the build dependencies
RUN dnf update -y
RUN dnf upgrade -y
RUN dnf install -y openjpeg2-devel
RUN dnf install -y ImageMagick-devel
RUN dnf install -y git
RUN dnf install -y gcc
RUN dnf install -y make
RUN dnf install -y tar
RUN dnf install -y findutils

# Installing GraphicsMagick is wholly unnecessary, but helps when using the
# build box for things like converting images.  Since opj2_encode doesn't
# support all formats, and ImageMagick has been iffy with some conversions for
# us, "gm convert" is a handy fallback.  As this is a multi-stage dockerfile,
# this installation doesn't hurt the final production docker image.
RUN dnf install -y GraphicsMagick

# Go comes after other installs to avoid re-pulling the more expensive
# dependencies when changing Go versions
RUN curl -L https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz > /tmp/go.tgz
RUN cd /opt && tar -xzf /tmp/go.tgz

# "Install" Go
RUN mkdir -p /usr/local/go
ENV GOPATH /usr/local/go
ENV GOROOT /opt/go
ENV PATH /opt/go/bin:/usr/local/go/bin:$PATH

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
ADD ./scripts /opt/rais-src/scripts
RUN make

# Manually build the ImageMagick plugin since we exclude it by default
RUN make bin/plugins/imagick-decoder.so

# Production image just installs runtime deps and copies in the binaries
FROM fedora:30 AS production
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Stolen from mariadb dockerfile: add our user and group first to make sure
# their IDs get assigned consistently
RUN groupadd -r rais && useradd -r -g rais rais

# Deps
RUN dnf update -y && dnf upgrade -y && dnf install -y openjpeg2 ImageMagick

ENV RAIS_TILEPATH /var/local/images
ENV RAIS_PLUGINS "*.so"
RUN touch /etc/rais.toml && chown rais:rais /etc/rais.toml
COPY --from=build /opt/rais-src/bin /opt/rais/

USER rais
EXPOSE 12415
ENTRYPOINT ["/opt/rais/rais-server"]
