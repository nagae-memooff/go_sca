#!/bin/bash
#
# Init Script to run ubeaver in daemon mode at boot time.
# chkconfig: - 67 67
# description: ubeaver daemon
# processname: ubeaver agent
#
# Run "/sbin/chkconfig --add ubeaver" to add the Run levels.
# This will setup the symlinks and set the process to run at boot.

# Source function library.
if [ -f /etc/rc.d/init.d/functions ]; then
  . /etc/rc.d/init.d/functions 
fi



#DIR=`pwd`
PROG=demo
DIR=/home/ewhine/deploy/$PROG
USER=ewhine
DAEMON=$DIR/$PROG
PIDFILE=/home/ewhine/var/run/$PROG.pid
STDOUT=/dev/null
STDERR=$DIR/$PROG.stderr.log
SDAEMON=$DIR/start-stop-daemon
ARGS="-c $DIR/$PROG.conf"
cd $DIR
#ulimit -n 100000

export GOGC=40

RETVAL=0

if [ ! -f $PIDFILE ]; then
  touch $PIDFILE
  chown ${USER}:${USER} $PIDFILE
fi

start () {
  PID=`pidof $PROG`
  if [ $? -eq 0 ]; then
    echo "$PROG has already been started!"
    exit 0;
  fi
  echo -n "Starting $PROG: "
  #start-stop-daemon --start --background --chuid $USER --chdir $DIR --make-pidfile --pidfile $PIDFILE --exec $DAEMON
  $SDAEMON --start --background --chuid $USER --chdir $DIR --make-pidfile --pidfile $PIDFILE --startas /bin/bash -- -c "exec $DAEMON $ARGS 1>>$STDOUT 2>>$STDERR"

  RETVAL=$?
  if [ $RETVAL -eq 0 ]; then
    sleep 1
    chown ${USER}:${USER} $PIDFILE
    success
  else
    failure
  fi
  echo
  return $RETVAL
}

stop () {
  status2 > /dev/null
  if [ $? -ne 0 ]; then
    echo "$PROG is not running!"
    return 0
  fi

  echo -n "Stopping $PROG: "
  killall $PROG > /dev/null 2>&1
  RETVAL=$?
  if [ $RETVAL -eq 0 ] ; then
    wait_for_quit
    success
  else
    failure
  fi
  echo
  return $RETVAL
}

status () {
  pid=`pidof $PROG`
  if [ $? -eq 0 ]; then
    echo "started"
  else
    echo "stoped"
  fi
}

status2 () {
  pid=`pidof $PROG`
  if [ $? -eq 0 ]; then
    echo "$PROG is running with pid $pid"
    return 0;
  else
    echo "$PROG is not running!"
    return 1;
  fi
  return 2
}

restart () {
  stop
  start
}

wait_for_quit () {
  status2 > /dev/null
  retval="$?"
  if [ $retval -eq 0 ]; then
    sleep 0.5
    wait_for_quit
  fi
}

cpu_total () {
  cat /proc/stat|head -n 1|awk '{for(i=2;i<=NF;i++)total+=$i;print total;}'
}

cpu_by_process() {
  pid="$1"
  cat /proc/$pid/stat|awk '{for(i=14;i<=15;i++)total+=$i; print total}'
}

cpu_usage() {
  cp1=`cpu_by_process $pid`
  ct1=`cpu_total`

  sleep 1
  cp2=`cpu_by_process $pid`
  ct2=`cpu_total`

  if [ "$ct2" == "$ct1" ]; then
    echo -n 0
    return
  fi

  echo -n $(echo $cp2, $cp1, $ct2, $ct1|awk '{printf "%0.2f", ($1-$2)*100/($3-$4)}')
}


mem_usage() {
  cat /proc/$pid/status|grep VmRSS|grep -v grep|awk '{printf "%.2f", $2/1024}'
}

info () {
  action=$1
  params=$2
  case "$action" in
    push)
      curl -f -d "params=$params" "http://127.0.0.1:10193/config_update_check" --silent  --connect-timeout 10 -m 30
      ret=$?
      exit $ret
      ;;
    pull)
      curl -f  "http://127.0.0.1:10193/get_info" --silent  --connect-timeout 10 -m 1
      ret=$?
      exit $ret
      ;;
  esac

}

conf_version () {
  conf_version_file="$DIR/conf/version"
  if [ -f "$conf_version_file" ]; then
    cat "$conf_version_file"
  else
    echo 
  fi
}

# See how we were called.
case "$1" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  status)
    status
    ;;
  restart|reload)
    restart
    ;;
  rusage)
    if [ ! -f $PIDFILE ]; then
      echo "0,0"
      exit 0
    fi

    pid=`cat $PIDFILE`
    if [ ! -f /proc/$pid/stat ]; then
      echo "0,0"
      exit 0
    fi

    comm=`cat /proc/$pid/comm`
    if [ "$comm" != "$PROG" ]; then
      echo "0,0"
      exit 0
    fi

    cpu_usage
    echo -n ","
    mem_usage
    ;;
  info)
    info "$2" "$3"
    ;;
  conf_version)
    conf_version
    ;;
  *)
    echo "Usage: $0 {start|stop|status|rusage|conf_version|info push json_msg|info pull}"
    exit 1
esac

exit 0
