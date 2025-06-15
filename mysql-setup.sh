#!/bin/bash

# MySQL Docker 管理脚本

set -e

COMPOSE_FILE="docker-compose.yml"

function usage() {
    echo "用法: $0 {start|stop|restart|status|logs}"
    echo "  start   - 启动MySQL容器"
    echo "  stop    - 停止MySQL容器"
    echo "  restart - 重启MySQL容器"
    echo "  status  - 查看容器状态"
    echo "  logs    - 查看容器日志"
    exit 1
}

function start_mysql() {
    echo "启动MySQL容器..."
    docker compose -f $COMPOSE_FILE up -d
    echo "MySQL容器已启动"
    echo "连接信息："
    echo "  主机: 127.0.0.1"
    echo "  端口: 3306"
    echo "  数据库: miniblog"
    echo "  用户名: miniblog"
    echo "  密码: miniblog1234"
}

function stop_mysql() {
    echo "停止MySQL容器..."
    docker compose -f $COMPOSE_FILE down
    echo "MySQL容器已停止"
}

function restart_mysql() {
    echo "重启MySQL容器..."
    docker compose -f $COMPOSE_FILE restart
    echo "MySQL容器已重启"
}

function status_mysql() {
    echo "MySQL容器状态："
    docker compose -f $COMPOSE_FILE ps
}

function logs_mysql() {
    echo "MySQL容器日志："
    docker compose -f $COMPOSE_FILE logs -f mysql
}

case "$1" in
    start)
        start_mysql
        ;;
    stop)
        stop_mysql
        ;;
    restart)
        restart_mysql
        ;;
    status)
        status_mysql
        ;;
    logs)
        logs_mysql
        ;;
    *)
        usage
        ;;
esac 