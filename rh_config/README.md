Red Hat RAIS config files
-----

Download and compile the go application as explained in the main
[readme](../README.md).  Then in your cloned repository, switch to this
directory (rh_config).  By default, the repository will be found at
`$GOPATH/src/github.com/uoregon-libraries/rais-image-server/`.

These actions should be performed as root (make sure you set GOPATH, as it
probably isn't set by default for root):

    mkdir /opt/chronam-support/
    cp init.sh /etc/init.d/rais
    cp rais.conf /etc
    # edit rais.conf
    cp $GOPATH/bin/rais-server /opt/chronam-support/rais-server
    chkconfig --add rais
    chkconfig rais on
    service rais start
