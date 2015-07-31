# Manual installation and usage

RAIS is easiest to run within a docker container.  The following instructions
may be useful for server admins who don't want docker on their server, but they
are pretty RHEL-centric.

## Setup

- *Optional*: Install openjpeg 2.1 (see below)
- Install imagemagick development files (`yum install ImageMagick-devel` on RHEL)
- [Install go](http://golang.org/doc/install)
- Set up the [`GOPATH` environment variable](http://golang.org/doc/code.html#GOPATH)
  - This tells go where to put the project
- Install the project without JP2 support:
  - `go get -u github.com/uoregon-libraries/rais-image-server/cmd/rais-server`
- *OR* install with JP2 support:
  - `go get -u -tags jp2 github.com/uoregon-libraries/rais-image-server/cmd/rais-server`

Please note that if you clone the repository from github, the makefile assumes
JP2 support in order to make development easier for me.

### Openjpeg installation

Openjpeg 2.1 must be installed to handle JP2 source images.  This is
**optional**, but if you have (or can build) tiled JP2 images, they will be
extremely space- and RAM-efficient, as well as being fairly fast.  In our
tests, TIFFs are slower with only a moderate load (a single user deliberately
panning and zooming quickly).

Installation depends on operating system, but we were able to rebuild the
Fedora SRPM for RedHat 6 and CentOS 7.

The general build algorthim is fairly straightforward:

- Go to http://koji.fedoraproject.org/koji/packageinfo?packageID=18369
- Grab the source rpm for the latest version of openjpeg2-2.1.x for the oldest
  version of fedora that exists
- Build on RHEL 6
- Install the openjpeg2 and openjpeg2-devel rpms which get built

We've specifically tested `openjpeg2-2.1.0-1`, but it stands to reason the
above steps will work for others as well.

Running RAIS
-----

### Manually

`$GOPATH/bin/rais-server --address=":8888" --tile-path="/path/to/data/batches"`

Note that if you wish to enable [IIIF](http://iiif.io/api/image/2.0/) support,
you must specify extra information on the command-line:

```bash
$GOPATH/bin/rais-server --address=":8888" --tile-path="/path/to/images" \
  --iiif-url="http://iiif.example.com/images/iiif" --iiif-tile-sizes="512,1024"
```

This would enable IIIF services with a base URL of `http://iiif.example.com:8888/images/iiif`.
Image info requests would then be at, e.g., `http://iiif.example.com:8888/images/iiif/myimage.jp2/info.json`.
It would report tile sizes of 512 and 1024, each with hard-coded scale factors
from 1 to 64 in powers of 2.  Currently the scale factors are not configurable.

Also note that the scheme and server (`http://my.iiifserver.example.com:8888`)
are informative for the IIIF information response, but aren't actually used by
the application otherwise.  IIIF information responses must include the full
URI to any given image.  The information must be correct, however, because IIIF
clients **will** use it to determine how to find the image tiles.

It is probably a good idea to set this up to run on server startup, and to
respawn if it dies unexpectedly:

### Red Hat 6 / 7

Read the provided [documentation for systems based on Red
Hat](rh_config/README.md).

Note that RHEL 7 uses a different system for init scripts (systemd), but what
we provide has worked on CentOS 7, so we're fairly confident it'll work on RHEL
7.  Ideally we'd have a proper systemd-based configuration - PRs would be most
appreciated here!

### Ubuntu

The original RAIS (previously known as "Brikker") was able to run on an Amazon
EC2 instance, but we haven't updated the config files with all the latest
changes.  We are no longer suggesting the old config files as there are too
many changes in the old brikker and RAIS.

PRs for working configs would be greatly appreciated.

