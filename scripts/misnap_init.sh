#!/bin/sh
# Copyright (C) gdy

luckydir=/data/lucky.daji
profile=/etc/profile

sed -i '/alias lucky=*/'d $profile
sed -i '/export luckydir=*/'d $profile
#h初始化环境变量
echo "alias lucky=\"$luckydir/lucky\"" >> $profile 
echo "export luckydir=\"$luckydir\"" >> $profile 

#设置init.d服务并启动lucky
ln -sf $luckydir/scripts/luckyservice /etc/init.d/lucky.daji
chmod 755 /etc/init.d/lucky.daji

log_file=`uci get system.@system[0].log_file`
i=0
while [ "$i" -lt 10 ];do
	sleep 3
	[ -n "$(grep 'init complete' $log_file)" ] && i=10 || i=$((i+1))
done
/etc/init.d/lucky.daji enable
/etc/init.d/lucky.daji start


