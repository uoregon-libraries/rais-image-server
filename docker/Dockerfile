# This generates a production image for RAIS with the S3 plugin enabled
#
# Examples:
#
#     # Simple case: just build the latest production image
#     docker build --rm -t uolibraries/rais:latest ./docker
#
#     # Build an older version of RAIS
#     docker build --rm -t uolibraries/rais:2.11.2 --build-arg gittag=v2.11.2 ./docker
#
#     # Generate the build image to simplify local development
#     docker build --rm -t uolibraries/rais:build --target build ./docker
FROM fedora:28 AS build
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

# Go comes after other installs to avoid re-pulling the more expensive
# dependencies when changing Go versions
RUN curl -L https://dl.google.com/go/go1.11.linux-amd64.tar.gz > /tmp/go.tgz
RUN cd /opt && tar -xzf /tmp/go.tgz

# "Install" Go
RUN mkdir -p /usr/local/go
ENV GOPATH /usr/local/go
ENV GOROOT /opt/go
ENV PATH /opt/go/bin:/usr/local/go/bin:$PATH

# Make sure the build box can lint code
RUN go get -u golang.org/x/lint/golint

# Clone the RAIS repo
RUN git clone https://github.com/uoregon-libraries/rais-image-server.git /opt/rais-src
WORKDIR /opt/rais-src

# Pull the latest RAIS master - that's always the stable release - unless an ENV var says otherwise
ARG gittag=master
RUN git checkout $gittag
RUN make

# Allow building older versions of RAIS by not failing if the s3 plugin doesn't exist
RUN go build -buildmode=plugin -o bin/plugins/s3-images.so rais/src/plugins/s3-images || true

# Production image just installs runtime deps and copies in the binaries
FROM fedora:28 AS production
LABEL maintainer="Jeremy Echols <jechols@uoregon.edu>"

# Stolen from mariadb dockerfile: add our user and group first to make sure
# their IDs get assigned consistently
RUN groupadd -r rais && useradd -r -g rais rais

# Deps
RUN dnf update -y && dnf upgrade -y && dnf install -y openjpeg2 ImageMagick

ENV RAIS_TILEPATH /var/local/images
RUN touch /etc/rais.toml && chown rais:rais /etc/rais.toml
COPY --from=build /opt/rais-src/bin /opt/rais/

USER rais
EXPOSE 12415
ENTRYPOINT ["/opt/rais/rais-server"]
