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

How efficient is RAIS?
-----

Very.

Unlike commercial solutions, RAIS is completely free and a breeze to test out
via Docker (see below).

RAIS utilizes the [OpenJPEG](https://github.com/uclouvain/openjpeg) libraries,
which have been improved over the years to deliver JP2 images at speeds
previously seen only in commercial applications.

[Historic Oregon Newspapers](https://oregonnews.uoregon.edu/) is our "flagship"
project using RAIS.  We receive anywhere from 2 to 4 million tile requests per
month, and response time is rarely above 500ms per tile.  It can handle dozens
of concurrent requests with minimal performance penalties.

As of mid-2018, we have nearly a million JP2 images, which take up roughly 4
terabytes and are mounted from external network storage.  And we don't run a
dozen servers with heavy layers of caching.  Our back-end is very modest:

- We have a single VM with 6 gigs of RAM
- Our server runs Solr, MariaDB, and [Open ONI](https://github.com/open-oni/open-oni) (a django application) in addition to RAIS
- We set up Apache to cache thumbnails (described below).  All other images are served from RAIS.
- We configured our instance of RAIS to cache up to 1000 tiles just in case
  many people are drawn to a single newspaper for any reason.

Despite our minimal hardware, RAIS uses roughly 600 megs of RAM, *including the
tile cache*.  Prior to adding tile-level caching, RAM usage was generally in
the range of 50-100 megs with brief spikes up to 400 megs during peak usage.
Despite the hefty costs of JP2 decoding and the number of incoming requests,
RAIS uses roughly 2-4 CPU hours per day, which is roughly equivalent to the
django stack's CPU usage.

Setup
-----

[Docker](https://www.docker.com/) is the preferred way to install and run RAIS.

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
[scripts/rundocker.sh](scripts/rundocker.sh).

On the first run, there will be a large download to get the image, but
after that it will be cached locally.

Once the container has been created, it can then be started and stopped via the
usual docker commands.

### Docker Demo

A demo-friendly `docker-compose.yml` file has been set up for convenience, and
uses nginx to serve up a web page that pulls image tiles from a proxied rais
backend.

You can run a quick demo simply by running `make docker` and then
`docker-compose up` from this project's root.  This may take a few minutes, but
once it's complete, you can visit `http://localhost` (or configure a custom URL
via `export URL=http://...` if you can't use "localhost" for some reason).
Click the IIIF link, and you'll be able to see the test JP2 in two different
forms using Open Seadragon: it will look the same, but its INFO response will
be slightly different from one to the other due to one of the two images having
an explicitly overridden info response.

Additionally, if you put other files into `docker/images`, the docker setup
scripts will automatically add them to the OSD tile source list, allowing you
to quickly test a variety of files.

Note that at the end of the file list, you'll likely get an error with the
image.  The final image in the demo is actually an external resource for
testing out the example plugin, external-images.  See the plugins section for
more details.

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
production image, will produce an image called "raisimageserver_rais-build"
which can be used to compile and run tests.  See
[docker/Dockerfile.build](docker/Dockerfile.build) for examples of how to make
this happen.  Also consider using [scripts/buildrun.sh](scripts/buildrun.sh) to ease compiling
and testing.  [scripts/dev.sh](scripts/dev.sh) is also available for easing the
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

Plugins
-----

### Example plugins

RAIS has a rudimentary plugin system in place.  Example plugins can be built
via `make plugins` or, if using docker, `./scripts/buildrun.sh make plugins`.

**Note**: the external images plugin is *not secure*.  It is meant for
demonstration only, and should not be used in a live server.  Remove
`bin/plugins/external-images.so` and restart RAIS if you accidentally ran it
with this plugin enabled.

The example Plugins may be built all at once (`make plugins`) or individually
(e.g., `make s3-images`).  Due to the security issues with the external images
plugin, it is strongly recommended you use `make plugins` only on development
servers that aren't exposed to the entire Internet.

#### External Images

The [example external-images plugin](src/plugins/external-images) showcases how
a plugin might be built to alter how RAIS interprets certain IDs.  The example
allows external images to be downloaded and converted to JP2s on the fly.
Though it would be an unusual use-case, this sort of plugin (with some tweaks)
could provide IIIF image features for external images you don't control, as
happens with third-party digital asset management systems.

Then start up the demo server and browse to the final image in the list.  After
15-30 seconds, you should see an image which was requested from hubblesite.org,
and which is now cached locally.

#### S3 Images

The [s3-images plugin](src/plugins/s3-images) is an example that can be tweaked
to pull from a particular S3 bucket on demand.  This plugin shows how you can
access configuration from a plugin as well as simple S3 access that's not
available from a "vanilla" RAIS build.

See [rais-example.toml](rais-example.toml) for configuration details.

Because this plugin is only able to access a server-configured S3 bucket, it is
considered to be secure, assuming your S3 bucket isn't able to get uploads from
unknown people on the Internet.

There is minimal documentation for this plugin at the moment (it was originally
just another example of how to *build* a plugin).  The comments at the top of
[main.go](src/plugins/s3-images/main.go) have more details.

#### DataDog

The [datadog plugin](src/plugins/datadog) shows the use of WrapHandler (see
below) by adding the DataDog tracing agent to all clients' requests.  This
allows high-level performance monitoring with minimal code.

This is again mostly an example (though it is production-ready if the behavior
makes sense for you), so the documentation is primarily contained within the
[main.go](src/plugins/datadog/main.go) file.

### Behavior

General behavior of plugins:

- Plugins are loaded from within a "plugins" directory.  "plugins" must be in
  the same location as the server binary.  The example plugin works in this
  way: when `make plugins` is run, the plugins are generated in
  `./bin/plugins/*.so`, whereas the server binary is `./bin/rais-server`.
- Plugin order matters.  Plugins will be processed in alphabetic order.  If you
  need plugins to have clear priorities, a naming scheme such as `00-name.so`
  may make sense (replacing `00` with a two-digit load order value).
- Plugins **must** be compiled on the same architecture that is used to compile
  the RAIS server.  This means the same server, the same openjpeg libraries,
  and the same version of Go.

### Plugin Functions

This section contains the complete list of exposed functions a plugin may
expose, and a detailed description of their function signature and behavior.

#### `SetLogger`

`func SetLogger(raisLogger *logger.Logger)`

All plugins may define this function, and there are no side-effects to worry
about when defining it (regarding ordering or multiple plugins "competing").
This function allows a plugin to make use of the central RAIS log subsystem so
that logged messages can be consistent.  Plugins don't have to expose this
function if they aren't logging any messages.

`logger.Logger` is defined by package `github.com/uoregon-libraries/gopkg/logger`.

#### `Initialize`

`func Initialize()`

All plugins may define this function.  This can be used to handle things like
custom configuration a plugin may need.  See the s3 plugin's Initialize method
for an example of that.

`Initialize()` is run after the logger is set up (unlike Go's internal `init()`
function), so you can safely use it.

#### `WrapHandler`

`func WrapHandler(pattern string, handler http.Handler) (http.Handler, error)`

WrapHandler is called three times in RAIS: once for the main IIIF handler, once
for the experimental DZI handler, and finally for the `/version` handler.  It
is meant only as a very high-level wrapper for the moment, and doesn't (easily)
allow adding custom handlers to RAIS.

A plugin implementing WrapHandler can use the pattern to identify the path
being wrapped -- though the IIIF path is variable, so this can be used for
logging or other "identity" logic, but not easily for filtering.  The handler
passed in is the current state of the handler.  A wrapper could add middleware,
logging, etc.  See the [datadog plugin](src/plugins/datadog) for an example.

Any number of plugins can implement WrapHandler.  Each plugin's returned
handler is sent to the next plugin in the list.

If a plugin handles this function, but needs to skip a particular pattern, it
should return `nil, plugins.ErrSkipped`.  This indicates to RAIS that the
plugin didn't fail, but simply chose to avoid trying to handle the given
pattern and handler.

#### `Teardown`

`func Teardown()`

All plugins may define a Teardown function for handling any necessary cleanup.
This is called when RAIS is about to exit, though it is not guaranteed to be
called (for instance, if power goes out or the server is force-killed).

#### `IDToPath`

`func IDToPath(id iiif.ID) (path string, err error)`

If the given ID isn't handled by this plugin, `plugins.ErrSkipped` should be
returned by the plugin (the `path` returned will be ignored).

The first plugin which returns a nil error "wins".  If there are multiple
plugins trying to convert IDs to paths, you must be sure you put them in a
sensible order.

IIIF Features
-----

RAIS must be run with a IIIF URL to support IIIF features, otherwise only the
chronam legacy handlers are active.  This is set via the RAIS configuration
value "IIIFURL" or simply on the command line:

    rais-server --iiif-url https://oregonnews.uoregon.edu/iiif

Other than possible bugs, we are ensuring we support IIIF Image API 2.1,
level 2.  We also support a handful of features beyond level 2.

Full list of features supported:

- Region:
  - "full"
  - "square"
  - "x,y,w,h": regionByPx
  - "pct:x,y,w,h": regionByPct
- Size:
  - "full"
  - "max"
  - "w," / sizeByW
  - ",h" / sizeByH
  - "pct:x" / sizeByPct
  - "w,h" / sizeByForcedWH
  - "!w,h" / sizeByWH
  - "sizeAboveFull"
  - "sizeByConfinedWh"
  - "sizeByDistortedWh"
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
  - gif (Note that this is VERY slow for some reason, and so is disabled by default)
- HTTP Features:
  - baseUriRedirect
  - cors
  - jsonldMediaType

To customize the capabilities for all images, a custom capabilities TOML file
can be specified on the command-line via `--capabilities-file [filename]`, the
config value `CapabilitiesFile`, or using the environment variable
`RAIS_CAPABILITIESFILE`.  You can remove undesired capabilities from the list
of what RAIS supports, which will prevent them from working if a client
requests them.  This can be helpful to avoid denial-of-service vectors, such as
the extremely slow GIF output (though this is now disabled by default).  See
[cap-max.toml](cap-max.toml) for an example that shows all currently supported
features.

When running as a IIIF server, you can browse to any valid Image's INFO page
to see the features supported.

To use a custom info.json response, you can create a file with the same name as
the JP2, with "-info.json" appended at the end.  e.g., `source.jp2-info.json`.
This can be useful for limiting features, custom resize values, etc.  To keep
the system working on any URL, you can set the `@id` value in the custom JSON
to `%ID%`.  Since IIIF ids are a full URL, changing paths, URLs, or ports will
break custom info.json files unless you allow the system to fill in the ID.
See [docker/images/testfile/test-world.jp2-info.json](docker/images/testfile/test-world.jp2-info.json)
for an example.

An example INFO URL, as clients like Open Seadragon expect, would look like
`http://example.com/iiif/source.jp2/info.json` (assuming, of course, that the
IIIF URL was specified as "http://example.com/iiif" and the file "source.jp2"
exists relative to the configured tile path).

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
are requested at a width and height of 1024 or below, in JPG format, can be
cached by setting TileCacheLen in /etc/rais.toml, or the RAIS_TILECACHELEN
environment variable.  This is disabled by default.

Setting the tile cache length to anything greater than zero will enable the
cache.

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

***Note***: *Tile caching is only recommended for systems with a small number
of images or systems that expect a lot of traffic to hit a small subset of the
collection, such as might be the case if there's a few featured images.
However, if you have an extra gig of RAM, it can be valuable to set a 1000-tile
cache even on large collections just to better handle an unexpected influx of
traffic to a small number of images, such as you might expect if part of your
collection gets featured in an online exhibit.*

Generating tiled, multi-resolution JP2s
---

Our newspaper JP2s are encoded in a way that makes them *very* friendly to
pan-and-zoom systems.  They are encoded as multi-resolution (think "zoom
levels"), tiled images.

JP2s that aren't encoded like this will not be nearly as memory- and
CPU-efficient.  We'd recommend tiling JP2s at a size of around 1024x1024.

### Open JPEG Tools

The openjpeg tools make this simple:

    /bin/opj2_compress -i input.tiff -o output.jp2 -t 1024,1024 -r 20.250 -n 6

Some notes:

- A rate [`-r`] of 20.250 is equivalent to a graphicsmagick "quality" of 70,
  which in JP2-land is about the same as a JPEG of quality 90-95 (very good).
- `-n 6` specifies that there are six resolution levels.  This can be optimized
  based on the image's size if desired, but 6 is the default.  Typically 6 will
  be fine, but a decent guideline is to start at 6 levels for a 16-megapixel
  image and add a resolution level each time the number of megapixels
  quadruples.  e.g., 16mp = `-n 6`, 64mp = `-n 7`, 256mp = `-n 8`, etc.
- You may have to build openjpeg tools manually to get support for converting
  some image formats, including TIFF and PNG

### Graphics Magick

If using graphics magick, encoding is still fairly easy, though note that you
may have to tweak the `jp2:numrlvls` argument depending on your images:

    gm convert input.tiff -flatten -quality 70 \
        -define jp2:prg=rlcp \
        -define jp2:numrlvls=7 \
        -define jp2:tilewidth=1024 \
        -define jp2:tileheight=1024 output.jp2

### Grayscale vs. color

Note that grayscale images will require one-third the memory and processing
power when compared to color images.  If your sources are grayscale, but you
scan in color for better preservation, consider building grayscale derivatives
for web display.

Known Limitations
-----

RAIS was built first and foremost to serve tiles for JP2s that always have multiple
resolution factors ("zoom levels") and are tiled.  It has been *amazing* for us
within that context, but there are areas where it won't perform well:

### JP2: Slow on huge files

Very large images (as in, hundreds of megapixels) can take a while to decode
tiles.  In some cases, 2-3 seconds per tile.  Unfortunately, this seems to be a
limitation of openjpeg.  If serving up files of this size, caching may be a
good idea.

### JP2: only supports RGB and Grayscale

YCC isn't supported directly (unless openjpeg does magic conversions for us,
which we haven't tested).  RGBa should work, but the alpha channel will be
ignored.  Embedded color profiles probably don't work, but they haven't been
tested.

### RAM usage should be monitored

Huge images and/or high traffic can cause the JP2 processor to chew up large
amounts of RAM.  If you use RAIS to serve up raw TIFF files, RAM usage is
likely to be a massive bottleneck.  The good news is that when handling tiled,
multi-resolution JP2s, RAM usage tends to be fairly predictable even under a
sizeable load.  And RAIS scales horizontally very well.

### IIIF Support isn't perfect

The IIIF support adheres to level 2 of the spec (as well as some extra
features), but it isn't as customizable as we would prefer.

When you don't provide your own info.json response (as described above), the
default response's quality choices are hard-coded to include color, gray, and
bitonal, even for gray/bitonal images.

It should also be noted that GIF output is amazingly slow.  Given that GIF
output isn't even a IIIF level 2 feature, we aren't planning to put much time
into troubleshooting the issue.  GIFs are available if you explicitly enable
them (via a capabilities file), but should be avoided except as one-offs.

### Poor performance for non-JP2 files

These aren't well-tested since our system is exclusively JP2.  Non-JP2 types
that are supported (TIFF, JPG, PNG, and GIF) have to be read in fully and then
cropped and resized within the application.  This will not be as fast as image
formats built for deep zooming (tiled JP2s for RAIS).

As an example: TIFF files are usually fast to process, but can potentially take
up a great deal of memory.  Sometimes this is okay, but it's a bottleneck
quickly when running a tiling server.  In our limited testing, TIFFs outperform
tiled JP2s only when load is extremely light.

License
-----

<img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />

RAIS Image Server is in the public domain under a
[CC0](http://creativecommons.org/publicdomain/zero/1.0/) license.
