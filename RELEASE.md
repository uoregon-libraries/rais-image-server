Release notes
=====

The top of this file contains the latest stable release and relevant notes
about what has changed since the previous release.

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
