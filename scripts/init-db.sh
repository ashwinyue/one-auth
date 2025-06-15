#!/bin/bash

# =====================================================
# One-Auth 数据库初始化脚本
# =====================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"
MYSQL_ROOT_PASSWORD="root123456"
MYSQL_USER="miniblog"
MYSQL_PASSWORD="miniblog1234"
MYSQL_DATABASE="miniblog"
DOCKER_CONTAINER="miniblog-mysql"
SQL_FILE="configs/miniblog.sql"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查文件是否存在
check_sql_file() {
    if [ ! -f "$SQL_FILE" ]; then
        print_error "SQL文件不存在: $SQL_FILE"
        exit 1
    fi
}

# 检查Docker容器是否运行
check_docker_container() {
    if ! docker ps | grep -q "$DOCKER_CONTAINER"; then
        print_warning "MySQL容器未运行，正在启动..."
        docker compose up -d mysql
        sleep 5
        
        if ! docker ps | grep -q "$DOCKER_CONTAINER"; then
            print_error "无法启动MySQL容器"
            exit 1
        fi
    fi
    print_success "MySQL容器运行正常"
}

# 使用Docker方式执行
execute_with_docker() {
    print_info "使用Docker方式执行SQL脚本..."
    
    check_docker_container
    check_sql_file
    
    # 执行SQL脚本
    if docker exec -i "$DOCKER_CONTAINER" mysql -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$SQL_FILE"; then
        print_success "数据库初始化完成！"
    else
        print_error "数据库初始化失败"
        exit 1
    fi
}

# 使用直接连接方式执行
execute_with_direct() {
    print_info "使用直接连接方式执行SQL脚本..."
    
    check_sql_file
    
    # 检查mysql命令是否可用
    if ! command -v mysql &> /dev/null; then
        print_error "mysql命令不可用，请安装MySQL客户端或使用Docker方式"
        exit 1
    fi
    
    # 执行SQL脚本
    if mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" < "$SQL_FILE"; then
        print_success "数据库初始化完成！"
    else
        print_error "数据库初始化失败"
        exit 1
    fi
}

# 交互式执行
execute_interactive() {
    print_info "启动交互式MySQL客户端..."
    
    check_docker_container
    
    print_info "连接到MySQL，可以手动执行以下命令："
    print_info "source /configs/miniblog.sql;"
    
    docker exec -it "$DOCKER_CONTAINER" mysql -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"
}

# 显示帮助信息
show_help() {
    echo "One-Auth 数据库初始化脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -d, --docker      使用Docker方式执行（推荐）"
    echo "  -c, --direct      使用直接连接方式执行"
    echo "  -i, --interactive 启动交互式MySQL客户端"
    echo "  -h, --help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -d             # 使用Docker方式初始化数据库"
    echo "  $0 --direct       # 使用直接连接方式初始化数据库"
    echo "  $0 -i             # 启动交互式客户端"
    echo ""
    echo "注意:"
    echo "  - 确保MySQL服务已启动: docker compose up -d mysql"
    echo "  - SQL文件位置: $SQL_FILE"
    echo "  - 数据库: $MYSQL_DATABASE"
}

# 主函数
main() {
    case "${1:-}" in
        -d|--docker)
            execute_with_docker
            ;;
        -c|--direct)
            execute_with_direct
            ;;
        -i|--interactive)
            execute_interactive
            ;;
        -h|--help)
            show_help
            ;;
        "")
            print_info "未指定执行方式，使用默认Docker方式"
            execute_with_docker
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 脚本开始
print_info "One-Auth 数据库初始化脚本启动"
print_info "当前工作目录: $(pwd)"

main "$@" 