#!/bin/bash
# -------------------------------------
# 服务启动脚本
#
# @author jesse.li
# @date 2016.05.13
# -------------------------------------

workDir=$(cd `dirname $0`; pwd)
binFile="$workDir/gopub"
pidFile="$workDir/gopub.pid"
error=""

cd $workDir

start() {
    nohup $binFile > /dev/null 2>&1 &
    echo $! > $pidFile
}

stop() {
    if [[ -e $pidFile ]]; then
        pid=`cat $pidFile`
        rm -f $pidFile
    else
        pid=`ps aux | grep gopub | grep -v grep | awk '{print $2}' | head -1`
    fi

    if [ "$pid"x != ""x ]; then
        kill -9 $pid
    else
        error="服务不在运行状态"
        return 1
    fi
}

case $1 in
start)
    if [[ -e $pidFile ]]; then
        echo "服务正在运行中, 进程ID: " $(cat $pidFile)
        exit 1
    fi
    echo -n "正在启动 ... "
    start
    sleep 1
    echo "成功, 进程ID:" $(cat $pidFile)
    ;;
stop)
    echo -n "正在停止 ... "
    stop
    if [[ $? -gt 0 ]]; then
        echo "失败, ${error}"
    else
        echo "成功"
    fi
    ;;
restart)
    echo -n "正在重启 ... "
    stop
    sleep 1
    start
    echo "成功, 进程ID:" $(cat $pidFile)
    ;;
status)
    if [[ -e $pidFile ]]; then
        pid=$(cat $pidFile)
    else
        pid=`ps aux | grep gopub | grep -v grep | awk '{print $2}' | head -1`
    fi
    if [[ -z "$pid" ]]; then
        echo "服务不在运行状态"
        exit 1
    fi
    exists=$(ps -ef | grep $pid | grep -v grep | wc -l)
    if [[ $exists -gt 0 ]]; then
        echo "服务正在运行中, 进程ID为${pid}"
    else
        echo "服务不在运行状态, 但进程ID文件存在"
    fi
    ;;
build)
    rebuild
    ;;
*)
    echo "GoPub启动脚本"
    echo "用法: "
    echo "    ./service.sh (start|stop|restart|status)"
;;
esac
