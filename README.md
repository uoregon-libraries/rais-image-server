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
  - `go get -u github.com/uoregon-libraries/newspaper-jp2-viewer/cmd/jp2tileserver`

### Openjpeg installation

Openjpeg 2.1 must be installed for this to work.  The previous approach which
used a checkout of the subversion repository is no longer supported.
Installation depends on operating system, but we were able to rebuild the
Fedora SRPM for RedHat 6.

The general build algorthim is fairly straightforward:

- Go to http://koji.fedoraproject.org/koji/packageinfo?packageID=18369
- Grab the source rpm for the latest version of openjpeg2-2.1.x for the oldest
  version of fedora that exists
- Build on RHEL 6
- Install the openjpeg2 and openjpeg2-devel rpms which get built

We've specifically tested `openjpeg2-2.1.0-1`, but it stands to reason the
above steps will work for others as well.

Running the tile server
-----

### Manually

`$GOPATH/bin/jp2tileserver --address=":8888" --tile-path="/path/to/data/batches"`

It is probably a good idea to set this up to run on server startup, and to
respawn if it dies unexpectedly.

### Red Hat 6

Read the provided [documentation for systems based on Red
Hat](rh_config/README.md).

### Ubuntu

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

Unfortunately, the version of chronam we're using has a lot of other dynamic
image URLs, so serving JP2s exclusively ended up requiring a lot of other
chronam hacks.  Our work isn't portable due to the extensive customizations we
have done to the site, but you can see the branch merge commit where we
centralized all dynamic image URLs [in this commit to the oregonnews
project](https://github.com/uoregon-libraries/oregonnews/commit/c8aad3287bf80cc4ca6716b91abd8b714be956a1)

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
