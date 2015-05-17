Release notes
=====

The top of this file contains the latest stable release and relevant notes
about what has changed since the previous release.

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
