Newspaper JP2 viewer
=======

The JP2 viewer was originally built by [eikeon](https://github.com/eikeon) as a
proof of concept of a tile server for JP2 images within
[chronam](https://github.com/LibraryOfCongress/chronam).  It has been updated
to allow more command-line options as well as a special command-line
application for verifying JP2 images can be served.  The University of Oregon's
primary use case is the [Historic Oregon
Newspapers](http://oregonnews.uoregon.edu/) project.

Known Limitations
-----

There are probably a lot of *unknown* limitations, so this should be considered
a very small list of what are likely a lot of problems.

- Resolution factors beyond 6 aren't supported

Very large images (as in, hundreds of megapixels) will have performance issues
as the server will have to manually resize anything smaller than 1/64th of the
original image.

- Resolution factors below 6 are barely supported

We couldn't figure out how to properly determine the number of resolution
factors a given image has.  When an image maxes out at res. factor 2, and a
request is made for something at factor 6, we try 6 and fail, then 5, fail, 4,
fail, 3, fail, 2, success.  This has a bit of overhead for each try, so it's
best to encode at a factor of 6 or else ensure you aren't requesting images
smaller than 1 / (2 ^ x) times the width/height of the image.

- Only supports RGB and Grayscale

And I'm not even sure how well we support different variations of those two
options... or even what variations might exist.  So... good luck?

- RAM usage could be ridiculous

We haven't even made a cursory attempt at curtailing RAM use.  Go's garbage
collection works well enough for our use case, but really large images and/or
lots of traffic could cause the system to easily chew up unreasonable amounts
of RAM.  Load testing is highly recommended.

- IIIF Support isn't perfect

The IIIF support adheres to level 1 of the spec, but it isn't as customizable
as we would prefer.  You can't specify per-image info.json responses; there is
no way to change the tile scale factors: 1, 2, 4, 8, 16, 32, and 64; and
there's no way to specify optimal resize targets.

IIIF viewers seem to pick moderately smart choices, but this server probably
won't work out of the box for, say, a 200+ megapixel image.

Setup
-----

- Install openjpeg 2.1 (see below)
- [Install go](http://golang.org/doc/install)
- Set up the [`GOPATH` environment variable](http://golang.org/doc/code.html#GOPATH)
  - This tells go where to put the project
- Install the project:
  - `go get -u github.com/uoregon-libraries/newspaper-jp2-viewer/cmd/jp2tileserver`

### Openjpeg installation

Openjpeg 2.1 must be installed for this to work.  The previous approach which
used a checkout of the subversion repository is no longer supported.
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

Running the tile server
-----

### Manually

`$GOPATH/bin/jp2tileserver --address=":8888" --tile-path="/path/to/data/batches"`

Note that if you wish to enable [IIIF](http://iiif.io/api/image/2.0/) support,
you must specify extra information on the command-line:

```bash
$GOPATH/bin/jp2tileserver --address=":8888" --tile-path="/path/to/images" \
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

Here's an example for running the original brikker on an Amazon EC2 instance:
[brikker-userdata.txt](https://gist.github.com/eikeon/5124717) (tested with
Ubuntu Server 12.10 from quick start).

*Note* that this was for an older version of the tile server and openjpeg.  The
scripts may need updates based on the latest information in this README.

Using with chronam
-----

To make this tile server work with
[chronam](https://github.com/LibraryOfCongress/chronam), you have two options.

You can [modify chronam directly](https://gist.github.com/eikeon/5124779),
which is easier for a quick test, but can make it tougher when chronam is
updated.

For a longer-term solution, you can instead make your web server proxy all
traffic for `/images/tiles/` to the tile server.  In Apache, you'd need to
enable proxy and proxy_http mods, and add this to your config:

`ProxyPass /images/tiles/ http://localhost:8888/images/tiles/`

Unfortunately, the version of chronam we're using has a lot of other dynamic
image URLs, so serving JP2s exclusively ended up requiring a lot of other
chronam hacks.  Our work isn't portable due to the extensive customizations we
have done to the site, but you can see the branch merge commit where we
centralized all dynamic image URLs [in this commit to the oregonnews
project](https://github.com/uoregon-libraries/oregonnews/commit/c8aad3287bf80cc4ca6716b91abd8b714be956a1)

Caching
-----

The server doesn't inherently cache the generated JPGs, which means every hit
will read the source JP2, extract tiles using openjpeg, and send them back to
the browser.  Depending on the amount of data and server horsepower, it may be
worth caching the tiles explicitly.

The server returns a valid Last-Modified header based on the last time the JP2
file changed, which Apache can use to create a simple disk-based cache:

```
CacheRoot /var/cache/apache2/mod_disk_cache
CacheEnable disk /images/tiles/
```

This won't be the smartest cache, but it will help in the case of a large
influx of people accessing the same newspaper.  It is highly advisable that the
`htcacheclean` tool be used in tandem with Apache cache directives, and it's
probably worth reading [the Apache caching
guide](http://httpd.apache.org/docs/2.2/caching.html).

License
-----

<img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />

The Newspaper JP2 Viewer is in the public domain under a
[CC0](http://creativecommons.org/publicdomain/zero/1.0/) license.
