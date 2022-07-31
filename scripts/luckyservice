#!/bin/sh /etc/rc.common
# Copyright (C) 2006-2011 OpenWrt.org

START=99
SERVICE_USE_PID=1
SERVICE_WRITE_PID=1
SERVICE_DAEMONIZE=1

#获取目录
DIR=$(cat /etc/profile | grep luckydir | awk -F "\"" '{print $2}')
[ -z "$DIR" ] && DIR=$(cat ~/.bashrc | grep luckydir | awk -F "\"" '{print $2}')
[ -z "$BINDIR" ] && BINDIR=$DIR


BIN=$BINDIR/lucky
CONF=$BINDIR/lucky.conf

start() {
	service_start $BIN -c $CONF &
}

stop() {
	service_stop $BIN
}

