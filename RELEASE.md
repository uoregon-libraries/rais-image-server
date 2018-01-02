Release notes
=====

The top of this file contains the latest stable release and relevant notes
about what has changed since the previous release.

Release 2.9.0
-----

- Adds a timeout for HTTP connections to (hopefully) avoid "accept4: too many
  open files" errors when things get too slow or networks drop temporarily
- Updates the build environment to use Go 1.9.2
- Improves the apache Open Seadragon / DeepZoom test container setup

Release 2.8.1
-----

- Fixes errors when a viewer uses URL query parameters

Release 2.8.0
-----

- Adds experimental support for handling the DeepZoom protocol

Release 2.7.2
-----

- Un-fixes tile caching to avoid huge delays when Open Seadragon pulls the initial thumbnails

This un-fix unfortunately adds a fair amount of potentially cached images which
may not be desired.  It's a necessary evil, however, as huge JP2s can take 30
seconds or more when Open Seadragon insists on pulling 5-10 lower-resolution
thumbnails before it will start working.

Release 2.7.1
-----

- Fixes tile caching to be significantly more selective about what is considered a tile

Release 2.7
-----

- Adds a new option to /etc/rais.toml for caching tiles in memory
- Uses Go 1.6.3 for Docker build

Release 2.6
-----

Major changes:

- Properly detects resolution levels in JP2 files, and reports scale factors accordingly
- Properly detects tile width and height, eliminating the need for manually specifying tiles on the command line
- Optionally caches data for info.json responses
- Adds the ability to override the IIIF info.json response per image
- Allows specifying configuration via /etc/rais.toml (see [rais-example.toml](rais-example.toml))
- Allows limiting RAIS features via an IIIF capabilities file (see [cap-max.toml](cap-max.toml)
  and [cap-level0.toml](cap-level0.toml) for examples)

Back-end improvements:

- Fixes init scripts for RHEL 6 users
- Uses Go 1.6 for the Docker build
- The build system now uses [gb](https://getgb.io/)
- Visiting the server URL + "/version" will report the current version of RAIS

Release 2.5
-----

- Adds docker support for production and development
- Removes the test which pulled huge JP2s from an external site

Release 2.4
-----

- Improves JP2 library detection when building with `-tags jp2`

Release 2.3
-----

- Fixes HUGE memory leak when handling pyramidal TIFFs

Release 2.2
-----

- Fixes bug with IIIF requests on non-JP2 sources

Release 2.1
-----

- Adds ImageMagick bindings for significantly faster TIFF decoding
- Makes JP2 support optional, off by default
- Allows chronam handler to use non-JP2 files
- Fixes a minor memory leak
- Removes annoying JP2 logging
- Makes it easier to register different backends for various image types

Release 2.0
-----

- Adds TIFF, JPG, PNG, and GIF support for source images (instead of just JP2)
- Adds PNG, GIF, and TIFF to output encoding options (instead of just JPG)
- Adds grayscale and bitonal ouput
- Adds force-resize and best-fit-resize options
- Adds mirroring support
- Dynamically determines compliance level for writing out info.json
- *Removes* legacy "info" handler
- *Removes* JP2 `Dimensions()` functionality (use `GetWidth` and `GetHeight` now)
- Now IIIF level 2 compliant
- Lots of formatting and lint fixes, and better testing

Release 1.0
-----

- Initial stable release under the RAIS name
- Initial stable release of IIIF features
