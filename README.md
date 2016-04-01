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

[Docker](https://www.docker.com/) is the preferred way to install and run RAIS.

See the [manual installation instructions](MANUAL-INSTALL.md) if you don't want to use
Docker, or you want to see exactly what's going on behind the scenes.

Note that specific build and production environments can be found in the docker
files and the Makefile's `docker` target, which may be useful for manual
installation.  [docker/README.md](docker/README.md) describes this in a little
more detail.

### Dockerhub

You can see an example in [rundocker.sh](rundocker.sh), but it looks like this:

```bash
docker run -d \
  --name rais \
  --privileged=true \
  -e PORT=12415 \
  -e IIIFURL="http://localhost:12415/iiif" \
  -e IIIFINFOCACHESIZE=10000
  -p 12415:12415 \
  -v $(pwd)/testfile:/var/local/images \
  uolibraries/rais
```

On the first run, there will be a large download to get the image, but
after that it will be cached locally.

Note that the environmental variables are optional, though IIIFURL will almost
certainly need to be changed in production:

- PORT: the port RAIS listens on, defaults to 12415
- IIIFURL: what RAIS reports as its server URL, defaults to localhost:$PORT/iiif
- IIIFINFOCACHESIZE is the number of items in the info cache.  Each item is
  very small, and 10,000 should use fewer than 5 megs of RAM.

Test by visiting `http://localhost:12415/iiif/test-world.jp2/full/full/0/default.jpg`,
then just configure the port/url/volume mount as needed.

Once the container has been created, it can then be started and stopped via the
usual docker commands.

### Build locally

You can clone the repository if you want to create your own RAIS image:

```bash
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server
make docker
```

*For contributors*: note that `make docker`, in addition to creating a
production image, will produce an image called "rais-build" which can be used
to compile and run tests.  See
[docker/Dockerfile.build](docker/Dockerfile.build) for examples of how to make
this happen.  Also consider using [buildrun.sh](buildrun.sh) to ease compiling
and testing.  [dev.sh](dev.sh) is also available for easing the
edit-compile-run loop on a system with no JP2 libraries, where compilation has
to go through docker.

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

Using with Open ONI
-----

RAIS works out of the box with [Open ONI](https://github.com/open-oni/open-oni),
a fork of chronam.  No hacking required!

IIIF Features
-----

When running as an IIIF server, you can browse to any valid Image's INFO page
to see the features supported.

To use a custom info.json response, you can create a file with the same name as
the JP2, with "-info.json" appended at the end.  e.g., `source.jp2-info.json`.
This can be useful for limiting features or reporting proper scale factors
(which are currently hard-coded), custom resize values, etc.  To keep the
system working on any URL, you can set the `@id` value in the custom JSON to
`%ID%`.  Since IIIF ids are a full URL, changing paths, URLs, or ports will
break custom info.json files unless you allow the system to fill in the ID.
See [testfile/info.json](testfile/info.json) for an example.

Other than possible bugs, we are ensuring we support level 2 at a minimum, as
well as a handful other features beyond level 2.

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
  - "w,h" / sizeByForcedWH
  - "!w,h" / sizeByWH
  - "sizeAboveFull"
- Rotation:
  - 0
  - "90,180,270" / rotationBy90s
  - "!0,!90,!180,!270" / mirroring
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

RAIS was built first and foremost to serve tiles for JP2s that always have exactly
six resolution factors ("zoom levels") and are tiled.  It has been *amazing* for us
within that context, but we don't know much about other uses, so outside of that
context, there may be issues worth consideration.

- JP2: resolution factors beyond 6 aren't supported

Very large images (as in, hundreds of megapixels) will have performance issues
as the server will have to manually resize anything smaller than 1/32nd of the
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

The IIIF support adheres to level 2 of the spec (as well as some extra
features), but it isn't as customizable as we would prefer.

- The info.json response makes assumptions in some cases that aren't optimal.
  You can override the responses with a custom info.json, but it's not automated.
- There is no way to change the tile scale factors: 1, 2, 4, 8, 16, 32.
- There's no way to specify optimal resize targets.
- The quality choices are hard-coded to include color, gray, and bitonal, even
  for gray/bitonal images

IIIF viewers seem to pick moderately smart choices, but this server may not
work out of the box for, say, a 200+ megapixel image.

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
CPU-efficient.  We'd recommend tiling JP2s at a size of around 1024x1024.  If
using graphics magick, a command like this can help:

    gm convert input.tiff -flatten -quality 70 \
        -define jp2:prg=rlcp \
        -define jp2:numrlvls=6 \
        -define jp2:tilewidth=1024 \
        -define jp2:tileheight=1024 output.jp2

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
up a great deal of memory.  Sometimes this is okay, but it's a bottleneck
quickly when running a tiling server.  In our limited testing, TIFFs outperform
tiled JP2s only when load is extremely light.

License
-----

<img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />

RAIS Image Server is in the public domain under a
[CC0](http://creativecommons.org/publicdomain/zero/1.0/) license.
