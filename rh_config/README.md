Red Hat jp2tileserver config files
-----

Download and compile the go application as explained in the main
[readme](../README.md).  Then in your cloned repository, switch to this
directory (rh_config).  By default, the repository will be found at
`$GOPATH/src/github.com/uoregon-libraries/newspaper-jp2-viewer/`.

These actions should be performed as root (make sure you set GOPATH, as it
probably isn't set by default for root):

    mkdir /opt/chronam-support/
    cp init.sh /etc/init.d/tileserver
    cp tileserver.conf /etc
    # edit tileserver.conf if you want to configure address, IIIF support, etc
    cp $GOPATH/bin/jp2tileserver /opt/chronam-support/jp2tileserver
    chkconfig --add tileserver
    chkconfig tileserver on
    service tileserver start
