brikker
=======

A server for generating and serving image tiles from a jp2 source image(s).

Example userdata for spinning up an ec2 instance running Brikker (tested with Ubuntu Server 12.10 from quick start):

[brikker-userdata.txt](https://gist.github.com/eikeon/5124717)

To use with [chronam](https://github.com/LibraryOfCongress/chronam) change your tile source to point at your brikker instance along the lines of:

[data-tile_url.diff](https://gist.github.com/eikeon/5124779)

with this setup Brikker is currently expecting jp2 files in the directory:

    /mnt

 in a location like:

    batch_dlc_jamaica_ver01/data/sn83030214/00175039259/1907100101/0002.jp2

