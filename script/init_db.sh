#!/bin/bash

# Prediction Relayer Service 数据库初始化脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="root"
DB_PASS=""
DB_NAME="prediction_relayer_service"

echo -e "${GREEN}=== Prediction Relayer Service 数据库初始化 ===${NC}"

# 检查 MySQL 是否可用
if ! command -v mysql &> /dev/null; then
    echo -e "${RED}错误: mysql 命令未找到，请先安装 MySQL 客户端${NC}"
    exit 1
fi

# 构建 MySQL 命令参数
MYSQL_CMD="mysql -h$DB_HOST -P$DB_PORT -u$DB_USER"
if [ -n "$DB_PASS" ]; then
    MYSQL_CMD="$MYSQL_CMD -p$DB_PASS"
fi

# 测试数据库连接
echo -e "${YELLOW}测试数据库连接...${NC}"
if $MYSQL_CMD -e "SELECT 1" &> /dev/null; then
    echo -e "${GREEN}数据库连接成功${NC}"
else
    echo -e "${RED}错误: 无法连接到数据库${NC}"
    echo "请检查数据库配置："
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  User: $DB_USER"
    exit 1
fi

# 创建数据库
echo -e "\n${YELLOW}创建数据库: $DB_NAME${NC}"
if $MYSQL_CMD -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"; then
    echo -e "${GREEN}数据库创建成功${NC}"
else
    echo -e "${RED}数据库创建失败${NC}"
    exit 1
fi

# 验证数据库
echo -e "\n${YELLOW}验证数据库...${NC}"
if $MYSQL_CMD -e "USE $DB_NAME; SELECT DATABASE();" &> /dev/null; then
    echo -e "${GREEN}数据库验证成功${NC}"
else
    echo -e "${RED}数据库验证失败${NC}"
    exit 1
fi

echo -e "\n${GREEN}=== 数据库初始化完成 ===${NC}"
echo -e "${YELLOW}注意: 表结构将通过 GORM AutoMigrate 自动创建${NC}"

