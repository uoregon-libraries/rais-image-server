[![Go Report Card](https://goreportcard.com/badge/github.com/uoregon-libraries/rais-image-server)](https://goreportcard.com/report/github.com/uoregon-libraries/rais-image-server)

RAIS Image Server
=======

RAIS was originally built by [eikeon](https://github.com/eikeon) as a 100% open
source, no-commercial-products-required, proof-of-concept tile server for JP2
images within [chronam](https://github.com/LibraryOfCongress/chronam).

It has been updated to allow more command-line options, more source file
formats, more features, conformance to the [IIIF](http://iiif.io/) spec, and
experimental DeepZoom support.

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

If you pull RAIS from [Dockerhub](https://hub.docker.com/r/uolibraries/rais/),
please note that you'll be getting the latest *stable* version.  It may not
have the same features as the development version.  You can look at the latest
stable version in github by browsing
[our master branch](https://github.com/uoregon-libraries/rais-image-server/tree/master).

For an example of running a docker image as a RAIS server, look at
[rundocker.sh](rundocker.sh).

On the first run, there will be a large download to get the image, but
after that it will be cached locally.

Once the container has been created, it can then be started and stopped via the
usual docker commands.

### Docker Demo

You can run a quick demo on a Linux system by pulling the "httpd:2.4" docker
image, running `./apache.sh` and then visiting `http://localhost`.  Click the
IIIF link, and you'll be able to see the test JP2 in two different forms using
Open Seadragon: it will look the same, but its INFO response will be slightly
different from one to the other due to one of the two images having an
explicitly overridden info response.

Additionally, if you put other files into `docker/images`, the `apache.sh`
script will automatically add them to the OSD tile source list, allowing you to
quickly test a variety of files.

Finally, you can test out the experimental DeepZoom support by choosing the
second link.

### Build locally

You can clone the repository if you want to create your own RAIS Docker image:

```bash
git clone https://github.com/uoregon-libraries/rais-image-server.git
cd rais-image-server
make docker
```

*For contributors*: note that `make docker`, in addition to creating a
production image, will produce an image called "rais-build:f27" which can be
used to compile and run tests.  See
[docker/Dockerfile.build](docker/Dockerfile.build) for examples of how to make
this happen.  Also consider using [buildrun.sh](buildrun.sh) to ease compiling
and testing.  [dev.sh](dev.sh) is also available for easing the
edit-compile-run loop on a system with no JP2 libraries, where compilation has
to go through docker.

Configuration
-----

RAIS uses a configuration system that allows environment variables, a config
file, and/or command-line flags.  See [rais-example.toml](rais-example.toml)
for an example of a configuration file.  RAIS will use a configuration
file if one exists at `/etc/rais.toml`.

The configuration file's values can be overridden by environment variables,
while command-line flags will override both configuration files and
environtmental variables.  Configuration is best explained and understood by
reading the example file above, which describes all the values in detail.

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
This can be useful for limiting features, custom resize values, etc.  To keep
the system working on any URL, you can set the `@id` value in the custom JSON
to `%ID%`.  Since IIIF ids are a full URL, changing paths, URLs, or ports will
break custom info.json files unless you allow the system to fill in the ID.
See [testfile/info.json](testfile/info.json) for an example.

To customize the capabilities for all images, a custom capabilities TOML file
can be specified on the command-line via `--capabilities-file [filename]`, the
config value `CapabilitiesFile`, or using the environment variable
`RAIS_CAPABILITIESFILE`.  You can remove undesired capabilities from the list
of what RAIS supports, which will prevent them from working if a client
requests them.  This can be helpful to avoid denial-of-service vectors, such as
the extremely slow GIF output.  See [cap-max.toml](cap-max.toml) for an example
that shows all currently supported features.

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

### info.json responses

We've implemented a simple LRU cache for info.json responses, which holds
10,000 entries by default.  The info.json data is very small, making this a
fairly efficient cache.  But the info.json data is very easy to generate, so
the value of caching is minimal, and may be removed in the future.

### Image responses

The server can optionally cache generated tiles under specific circumstances,
but doesn't inherently cache the other images such as thumbnails.  Tiles which
are requested at a width of 1024, 512, or 256, in JPG format, can be cached by
setting TileCacheLen in /etc/rais.toml, or the RAIS_TILECACHELEN environment
variable.  This defaults to not be enabled due to the cost of caching tiles
compared to the benefits, particularly on large collections of images.

Setting the tile cache length to anything greater than zero will enable the
cache.  This is only recommended for systems with a small number of images or
systems that expect a lot of traffic to hit a small subset of the collection,
such as might be the case if there's a few featured images.

For resize requests such as thumbnails, caching can prove very beneficial, but
for now RAIS doesn't have resize-only caching, as Apache handles this for us
very effectively.

The server returns a valid Last-Modified header based on the last time the JP2
file changed, which Apache can use to create a simple disk-based cache:

```
# Cache thumbnails (and only thumbnails)
CacheRoot /var/cache/httpd/mod_disk_cache
CacheEnable disk /images/resize

# Allow a total of 4096 content directories at two levels so we never have
# more than 64 directories in any other directory.  If we cache a million
# thumbnails, we'll still only end up with about 250 files per content
# directory.
CacheDirLength 1
CacheDirLevels 2

# Change !RAIS_HOST! below to serve tiles and thumbnails from RAIS
AllowEncodedSlashes NoDecode
ProxyPassMatch ^/images/resize/([^/]*)/full/([0-6][0-9][0-9],.*jpg)$ http://!RAIS_HOST!:12415/images/iiif/$1/full/$2 nocanon
ProxyPassMatch ^/images/iiif/(.*(jpg|info\.json))$ http://!RAIS_HOST!:12415/images/iiif/$1 nocanon
```

This kind of configuration allows us to split the resize requests from other
requests in order to have a reasonably intelligent disk cache, which is kept
separate from the in-memory tile cache.

This won't be the smartest cache, but it will help when search results pages
are used on large collections.  It is highly advisable that the `htcacheclean`
tool be used in tandem with Apache cache directives, and it's probably worth
reading [the Apache caching guide](http://httpd.apache.org/docs/2.2/caching.html).

Known Limitations
-----

RAIS was built first and foremost to serve tiles for JP2s that always have exactly
six resolution factors ("zoom levels") and are tiled.  It has been *amazing* for us
within that context, but we don't know much about other uses, so outside of that
context, there may be issues worth consideration.

### JP2: Slow on huge files

Very large images (as in, hundreds of megapixels) can take a while to decode
tiles.  In some cases, 2-3 seconds per tile.  Unfortunately, this seems to be a
limitation of openjpeg.  If serving up files of this size, external caching may
be a good idea.

### JP2: only supports RGB and Grayscale

YCC isn't supported directly (unless openjpeg does magic conversions for us,
which we haven't tested).  RGBa should work, but the alpha channel will be
ignored.  Embedded color profiles probably don't work, but they haven't been
tested.

### RAM usage should be monitored

Huge images and/or high traffic can cause the JP2 processor to chew up large
amounts of RAM.  The good news is that since compiling RAIS under Go 1.6, our
RAM is significantly lower and more predictable than with Go 1.4.

Stats from about two months of monitoring:

- Under Go 1.4, RAIS would slowly grow in RAM use until it was routinely above
  1 gig of RAM (even when under relatively low load), with spikes above 2 gigs
- Under Go 1.6, RAIS is typically under 80 megs of RAM, with spikes being few
  and far between, with the worst spike just over 400 megs

(For reference, RAIS serves about 800,000 tiles a week)

**Please note**: if you enable tile caching, RAM usage will increase, possibly
by a very significant amount.

### IIIF Support isn't perfect

The IIIF support adheres to level 2 of the spec (as well as some extra
features), but it isn't as customizable as we would prefer.

When you don't provide your own info.json response (as described above), the
default response's quality choices are hard-coded to include color, gray, and
bitonal, even for gray/bitonal images.

It should also be noted that GIF output is amazingly slow.  Given that GIF
output isn't even an IIIF level 2 feature, we aren't planning to put much time
into troubleshooting the issue.  GIFs are available, but should be avoided
except as one-offs.

### Not all JP2 files are created equally

Our newspaper JP2s are encoded in a way that makes them *very* friendly to
pan-and-zoom systems.  They are encoded with tiling, which allows pieces of
the JP2 to be read independently, and significantly reduces the memory needed
to serve the data up to a viewer on the fly.

JP2s that aren't encoded like this will not be nearly as memory- and
CPU-efficient.  We'd recommend tiling JP2s at a size of around 1024x1024.  If
using graphics magick, a command like this can help:

    gm convert input.tiff -flatten -quality 70 \
        -define jp2:prg=rlcp \
        -define jp2:numrlvls=7 \
        -define jp2:tilewidth=1024 \
        -define jp2:tileheight=1024 output.jp2

Additionally, grayscale images will require one-third the memory and processing
power when compared to color images.  If your sources are grayscale, but you
encode to RGB for better preservation, consider building grayscale derivatives
for web display.

### Poor performance for non-JP2 files

These aren't well-tested since our system is exclusively JP2.  Non-JP2 types
that are supported (TIFF, JPG, PNG, and GIF) have to be read in fully and then
cropped and resized in Go.  This will not be as fast as image formats built for
deep zooming (tiled JP2s for RAIS).

As an example: TIFF files are usually fast to process, but can potentially take
up a great deal of memory.  Sometimes this is okay, but it's a bottleneck
quickly when running a tiling server.  In our limited testing, TIFFs outperform
tiled JP2s only when load is extremely light.

License
-----

<img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />

RAIS Image Server is in the public domain under a
[CC0](http://creativecommons.org/publicdomain/zero/1.0/) license.
