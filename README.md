Newspaper JP2 viewer
=======

The JP2 viewer was originally built by [eikeon](https://github.com/eikeon) as a
proof of concept of a tile server for JP2 images within
[chronam](https://github.com/LibraryOfCongress/chronam).  It has been updated
to allow more command-line options as well as a special command-line
application for verifying JP2 images can be served.  The University of Oregon's
primary use case is the [Historic Oregon
Newspapers](http://oregonnews.uoregon.edu/) project.

Technically the jp2 viewer could be used for a variety of systems, but is built
with a few [chronam](https://github.com/LibraryOfCongress/chronam) rules
hard-coded.  It may be worthwhile to consider making it more configurable, but
that's not our goal at this time.

Setup
-----

- Install cmake 2.8 or later
- Install subversion
- Install openjpeg *from source* (see below)
- [Install go](http://golang.org/doc/install)
- Set up the [`GOPATH` environment variable](http://golang.org/doc/code.html#GOPATH)
  - This tells go where to put the project
- Install the project:
  - `go get -u github.com/uoregon-libraries/newspaper-jp2-viewer`
  - `go install github.com/uoregon-libraries/newspaper-jp2-viewer/cmd/jp2tileserver`

### Openjpeg installation

Openjpeg has to be installed from source until a new 2.x release is available.
2.0 will not work.  The following bash checks out the latest version of 2.0
from trunk (this project is currently hard-coded to use 2.0, and the bleeding
edge openjpeg code is set to be version 2.1):

```bash
cd /usr/local/src
svn checkout http://openjpeg.googlecode.com/svn/trunk/ openjpeg
cd openjpeg
svn update -r 2722
cmake .                     # This might be "cmake28 ."
sudo make install
sudo ldconfig
```

Running the tile server
-----

`$GOPATH/bin/jp2tileserver --address=":8888" --tile-path="/path/to/data/batches"`

It is probably a good idea to set this up to run on server startup, and to
respawn if it dies unexpectedly.

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
[CCO](http://creativecommons.org/publicdomain/zero/1.0/") license.
