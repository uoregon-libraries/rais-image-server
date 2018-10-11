[![Go Report Card](https://goreportcard.com/badge/github.com/uoregon-libraries/rais-image-server)](https://goreportcard.com/report/github.com/uoregon-libraries/rais-image-server)

RAIS Image Server
=======

RAIS was originally built by [eikeon](https://github.com/eikeon) as a 100% open
source, no-commercial-products-required, proof-of-concept tile server for JP2
images within [chronam](https://github.com/LibraryOfCongress/chronam).

It has been updated to allow more command-line options, more source file
formats, more features, conformance to the [IIIF](http://iiif.io/) spec, and
experimental DeepZoom support.

RAIS is very efficient, completely free, and easy to set up and run.  See our
[wiki](https://github.com/uoregon-libraries/rais-image-server/wiki) pages for
more details and documentation.

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

IIIF Features
-----

Other than possible bugs, we are ensuring we support IIIF Image API 2.1,
level 2.  We also support a handful of features beyond level 2.

See [the IIIF Features wiki page](https://github.com/uoregon-libraries/rais-image-server/wiki/IIIF-Features)
for an in-depth look at feature support.

Caching
-----

RAIS can internally cache the IIIF `info.json` requests, individual tile
requests, and, if it's in use, S3 images to a locally configured location.  See
the [RAIS Caching](https://github.com/uoregon-libraries/rais-image-server/wiki/Caching)
wiki page for details.

Generating tiled, multi-resolution JP2s
---

RAIS performs best with JP2s which are generated as tiled, multi-resolution
(think "zoom levels") images.  Generating images like this is fairly easy with
either the openjpeg tools or graphicsmagick.  Other tools probably do this
well, but we've only directly used those.

You can find detailed instructions on the
[How to encode jp2s](https://github.com/uoregon-libraries/rais-image-server/wiki/How-To-Encode-JP2s)
wiki page.

License
-----

<img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />

RAIS Image Server is in the public domain under a
[CC0](http://creativecommons.org/publicdomain/zero/1.0/) license.
