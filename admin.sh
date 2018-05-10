#!/bin/bash
export PATH="/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin:/root/bin:"

SERVER="nightwatch"
BASE_DIR=$PWD
LOG=$BASE_DIR/logs
mkdir -p $LOG

function start()
{
	if [ "`pgrep $SERVER -u $UID`" != "" ];then
		echo "$SERVER already running"
		exit 1
	fi

	nohup $BASE_DIR/$SERVER -log_dir=logs server &>$LOG/${SERVER}_$(date +%F).log &

	echo "sleeping..." &&  sleep 3
	$BASE_DIR/$SERVER register sample.yaml

	# check status
	if [ "`pgrep $SERVER -u $UID`" == "" ];then
		echo "$SERVER start failed"
		exit 1
	fi
}

function status() 
{
	if [ "`pgrep $SERVER -u $UID`" != "" ];then
		echo $SERVER is running
	else
		echo $SERVER is not running
	fi
}

function stop() 
{
	if [ "`pgrep $SERVER -u $UID`" != "" ];then
		kill `pgrep $SERVER -u $UID`
	fi

	echo "sleeping..." &&  sleep 3

	if [ "`pgrep $SERVER -u $UID`" != "" ];then
		echo "$SERVER stop failed"
		exit 1
	fi
}

case "$1" in
	'start')
	start
	;;  
	'stop')
	stop
	;;  
	'status')
	status
	;;  
	'restart')
	stop && start
	;;  
	*)  
	echo "usage: $0 {start|stop|restart|status}"
	exit 1
	;;  
esac

