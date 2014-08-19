verifyJP2s
=====

This command-line utility can be useful for a sanity test against your JP2s and
the versions of the JP2 viewer and openjpeg installed.

Installation
-----

Follow the directions for [installing the project](/README.md), as the dependencies
are the same.

After it's installed, run `go install
github.com/uoregon-libraries/newspaper-jp2-viewer/cmd/...`.  All commands in
the project will be installed at `$GOPATH/bin`.

Usage
-----

Create a file containing full paths to your JP2 files.  Call it "files.txt".
Run `$GOPATH/bin/verifyJP2s files.txt`.  Watch the output.  If you anything but
"SUCCESS" between the hyphenated lines, something didn't work.

Example Output
-----

This is from a select list of chronam's sample newspaper images:

```
Attempting to read from files.txt
BEGIN
---
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0010.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0036.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0054.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0049.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0026.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0016.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0021.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0034.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0014.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0027.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0042.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0032.jp2"
SUCCESS: "/opt/chronam/data/batches/batch_uuml_thys_ver01/data/sn83045396/print/1911091701/0044.jp2"
---
COMPLETE
```
