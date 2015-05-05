#!/bin/sh
#
# tileserver This shell script takes care of starting and stopping the tile
#   server for newspaper JP2 images
#
# chkconfig:    - 80 20
# description:  Tile server for newspaper JP2 images
# processname:  tileserver
# pidfile:      /var/run/tileserver.pid

### BEGIN INIT INFO
# Provides:           tileserver
# Required-Start:     $local_fs $remote_fs $network $named $time
# Required-Stop:      $local_fs $remote_fs $network $named $time
# Short-Description:  Start and stop jp2 tile server
# Description:        tileserver serves tiles as JPG images for front-end
#                     viewers to present a "pan & zoom" view of specially-
#                     formatted JP2 images
### END INIT INFO

# Source function library.
. /etc/rc.d/init.d/functions

name=`basename $0`
pid_file="/var/run/$name.pid"
stdout_log="/var/log/$name.log"
stderr_log="/var/log/$name.err"

prog="jp2tileserver"
exec="/opt/chronam-support/$prog"
cmd="$exec --tile-path=/opt/chronam/data/batches --address=:8080"

restartfile=/tmp/$prog.restart
lockfile=/var/lock/subsys/$prog

loop_tileserver() {
  # Until this file is gone, we want to restart the process
  touch $restartfile
  retry=5

  while [ -f $restartfile ] && [ $retry -gt 0 ]; do
    laststart=`date +%s`
    $cmd >>$stdout_log 2>>$stderr_log

    newdate=`date +%s`
    let timediff=$newdate-$laststart

    # Log the restart to stderr and stdout logs in an apache-like format
    if [-f $restartfile ]; then
      local logdate=`date +"[%a %b %d %H:%M:%S %Y]"`
      local message="Restarting server, ran for $timediff seconds before error"
      echo "$logdate [WARN] $message" >> $stdout_log
      echo "$logdate [WARN] $message" >> $stderr_log
    fi

    # Reset the retry counter as long as we don't restart too often; otherwise
    # we break out of the loop and assume we have a major failure
    let retry=$retry-1
    [ $timediff -gt 5 ] && retry=5
  done

  [ $retry -eq 0 ] && echo "Restart loop detected, aborting $prog" >> $stderr_log && exit 1
}

# Returns "true" (zero) if the passed-in app is found
is_running() {
  ps -C $1 >/dev/null 2>/dev/null || return 1
  return 0
}

wait_for_pid() {
  delay=5
  while [ $delay -gt 0 ] && [ -z "$pid" ]; do
    if is_running $prog; then
      pid=`pidof $prog`
      return 0
    fi
    sleep 1
    let delay=$delay-1
  done
  return 1
}

start() {
    [ -x $exec ] || exit 5

    echo -n $"Starting $prog: "

    # Loop the command
    loop_tileserver &

    # Try to find the pid
    pid=
    wait_for_pid $prog

    if [ -z "$pid" ]; then
      failure && echo
      return 1
    fi

    echo $pid > $pid_file
    touch $lockfile
    success && echo
    return 0
}

stop() {
    # Don't let the loop continue when we kill the process
    rm -f $restartfile

    echo -n $"Stopping $prog: "
    killproc $prog
    retval=$?
    echo
    [ $retval -eq 0 ] && rm -f $lockfile
    return $retval
}

restart() {
    stop
    start
}

reload() {
    restart
}

force_reload() {
    restart
}

rh_status() {
    # run checks to determine if the service is running or use generic status
    status $prog
}

rh_status_q() {
    rh_status >/dev/null 2>&1
}


case "$1" in
    start)
        rh_status_q && exit 0
        $1
        ;;
    stop)
        rh_status_q || exit 0
        $1
        ;;
    restart)
        $1
        ;;
    reload)
        rh_status_q || exit 7
        $1
        ;;
    force-reload)
        force_reload
        ;;
    status)
        rh_status
        ;;
    condrestart|try-restart)
        rh_status_q || exit 0
        restart
        ;;
    *)
        echo $"Usage: $0 {start|stop|status|restart|condrestart|try-restart|reload|force-reload}"
        exit 2
esac
exit $?
