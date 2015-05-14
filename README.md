RAIS Image Server
=======

RAIS was originally built by [eikeon](https://github.com/eikeon) as a 100% open
source, no-commercial-products-required, proof-of-concept tile server for JP2
images within [chronam](https://github.com/LibraryOfCongress/chronam).

It has been updated to allow more command-line options, more source file
formats, more features, and conformance to the [IIIF](http://iiif.io/) spec.

The University of Oregon's primary use case is the [Historic Oregon
Newspapers](http://oregonnews.uoregon.edu/) project.

Setup
-----

- Install openjpeg 2.1 (see below)
- [Install go](http://golang.org/doc/install)
- Set up the [`GOPATH` environment variable](http://golang.org/doc/code.html#GOPATH)
  - This tells go where to put the project
- Install the project:
  - `go get -u github.com/uoregon-libraries/rais-image-server/cmd/rais-server`

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

IIIF Features
-----

When running as an IIIF server, you can browse to any valid Image's INFO page
to see the features supported.  At the moment, there is no smart per-image
feature support.  Other than possible bugs, we are ensuring we support level 1
at a minimum, as well as a handful other features beyond level 1.

An example INFO request would look like `http://example.com/iiif/source.jp2/info.json`,
assuming your server is at `example.com`, the IIIF prefix is `iiif`, and the
file "source.jp2" exists relative to the configured tile path.

Full list of features supported:

- Region:
  - "full"
  - "x,y,w,h": regionByPx
  - "pct:x,y,w,h": regionByPct
- Size:
  - "full"
  - "w," / sizeByW
  - ",h" / sizeByH
  - "pct:x" / sizeByPct
  - "sizeAboveFull"
- Rotation:
  - 0
  - "90,180,270" / rotationBy90s
- Quality:
  - "default"
  - "native" (same as "default")
  - "color"
  - "gray"
  - "bitonal"
- Format:
  - jpg (This is the best format for a speedy encode and small download)
  - png
  - tif
  - gif (Note that this is VERY slow for some reason!)
- HTTP Features:
  - baseUriRedirect
  - cors
  - jsonldMediaType

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

Known Limitations
-----

There are probably a lot of *unknown* limitations, so this should be considered
a very small list of what are likely a lot of problems.

- JP2: resolution factors beyond 6 aren't supported

Very large images (as in, hundreds of megapixels) will have performance issues
as the server will have to manually resize anything smaller than 1/64th of the
original image.

- JP2: resolution factors below 6 are barely supported

We couldn't figure out how to properly determine the number of resolution
factors a given image has.  When an image maxes out at res. factor 2, and a
request is made for something at factor 6, we try 6 and fail, then 5, fail, 4,
fail, 3, fail, 2, success.  This has a bit of overhead for each try, so it's
best to encode at a factor of 6 or else ensure you aren't requesting images
smaller than 1 / (2 ^ x) times the width/height of the image.

- JP2: only supports RGB and Grayscale

And I'm not even sure how well we support different variations of those two
options... or even what variations might exist.  So... good luck?

- RAM usage could be ridiculous

We haven't even made a cursory attempt at curtailing RAM use.  Go's garbage
collection works well enough for our use case, but really large images and/or
lots of traffic could cause the system to easily chew up unreasonable amounts
of RAM.  Load testing is highly recommended.

- IIIF Support isn't perfect

The IIIF support adheres to level 1 of the spec (as well as some extra
features), but it isn't as customizable as we would prefer.  You can't specify
per-image info.json responses; there is no way to change the tile scale
factors: 1, 2, 4, 8, 16, 32, and 64; and there's no way to specify optimal
resize targets.

IIIF viewers seem to pick moderately smart choices, but this server probably
won't work out of the box for, say, a 200+ megapixel image.

Since there are no per-image info.json responses, the quality choices are, to
some degree, incorrect.  A grayscale image will report it has a color, gray,
and bitonal qualities available, when in fact it only has gray and bitonal.

It should also be noted that GIF output is amazingly slow.  Given that GIF
output isn't even an IIIF level 2 feature, we aren't planning to put much time
into troubleshooting the issue.  GIFs are available, but not likely to be used
except as one-offs.

- Not all JP2 files are created equally

Our newspaper JP2s are encoded in a way that makes them *very* friendly to
pan-and-zoom systems.  They are encoded with tiling, which allows pieces of
the JP2 to be read independently, and significantly reduces the memory needed
to serve the data up to a viewer on the fly.

JP2s that aren't encoded like this will not be nearly as memory- and
CPU-efficient.  We'd recommend tiling JP2s at a size of around 1024x1024.

Additionally, grayscale images will require one-third the memory and processing
power when compared to color images.  If your sources are grayscale, but you
encode to RGB for better preservation, consider building grayscale derivatives
for web display.

- Unknown performance metrics for non-JP2 files

These aren't well-tested since our system is exclusively JP2.  Non-JP2 types
that are supported (TIFF, JPG, PNG, and GIF) have to be read in fully and then
cropped and resized in Go.  This will not be as fast as image formats built for
deep zooming and run under a high-performance image server such as [IIP
Image](http://iipimage.sourceforge.net/).

As an example: TIFF files are usually fast to process, but can potentially take
up a great deal of memory.  In some cases, the speed outweighs the memory
costs, as the decoding happens so fast the RAM is able to be freed before it
becomes a problem.  Again, load-testing is extremely important here.

License
-----

<img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />

RAIS Image Server is in the public domain under a
[CC0](http://creativecommons.org/publicdomain/zero/1.0/) license.
