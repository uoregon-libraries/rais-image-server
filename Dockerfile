FROM fedora:22
MAINTAINER Jeremy Echols <jechols@uoregon.edu>

# Try to put the slowest stuff up top so the docker cache isn't fried by
# changing things like RAIS port setting

# Install all the dependencies - this will take a while, so we break up the
# packages to one per line.  If we need to add something, it's not gonna bust
# the whole docker cache.
RUN dnf install -y openjpeg2-devel
RUN dnf install -y golang
RUN dnf install -y ImageMagick-devel
RUN dnf install -y git

# Grab the RAIS repo and compile it with JP2 support
ENV GOPATH /tmp/go
RUN go get -u -tags jp2 github.com/uoregon-libraries/rais-image-server/cmd/rais-server

RUN mkdir /opt/rais
RUN cp /tmp/go/bin/rais-server /opt/rais

ENV PORT 12415
ENV TILESIZES 512
ENV IIIFURL http://localhost:$PORT/iiif

EXPOSE $PORT
ENTRYPOINT /opt/rais/rais-server --iiif-url $IIIFURL --address ":$PORT" --iiif-tile-sizes "$TILESIZES" --tile-path /var/local/images
