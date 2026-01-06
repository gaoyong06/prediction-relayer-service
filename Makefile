# Prediction Relayer Service Makefile
# 使用 devops-tools 的通用 Makefile

SERVICE_NAME=prediction-relayer-service
API_PROTO_PATH=api/relayer/v1/relayer.proto
API_PROTO_DIR=api/relayer/v1

# 服务特定配置
SERVICE_DISPLAY_NAME=Prediction Relayer Service
HTTP_PORT=8119
GRPC_PORT=9119
TEST_CONFIG=test/api/api-test-config.yaml
CONF_PROTO_PATH=internal/conf/config.proto
RUN_MODE=debug

# 构建配置
BUILD_OUTPUT=./bin/relayer-service
RUN_MAIN=cmd/server/main.go cmd/server/wire_gen.go

# 测试数据库配置
TEST_DB_HOST ?= 127.0.0.1
TEST_DB_PORT ?= 3306
TEST_DB_USER ?= root
TEST_DB_NAME ?= prediction_relayer_service

MYSQL_CMD = mysql -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u $(TEST_DB_USER) -D $(TEST_DB_NAME)
ifneq ($(TEST_DB_PASSWORD),)
MYSQL_CMD += -p$(TEST_DB_PASSWORD)
endif

# 测试数据清理命令
TEST_CLEAN_DATA_CMD = $(MYSQL_CMD) -e 'DELETE FROM transactions WHERE task_id LIKE '\''test-%'\''; DELETE FROM builder_fees WHERE builder_api_key LIKE '\''test-%'\'';' && redis-cli -h 127.0.0.1 -p 6379 FLUSHDB > /dev/null 2>&1 || true

# 清理测试数据
.PHONY: clean-test-data
clean-test-data:
	@echo "清理测试数据..."
	@$(TEST_CLEAN_DATA_CMD)
	@echo "测试数据清理完成"

# 引入通用 Makefile
DEVOPS_TOOLS_DIR := $(shell cd .. && pwd)/devops-tools
include $(DEVOPS_TOOLS_DIR)/Makefile.common

# 服务特定的目标（如果需要覆盖通用目标，在这里定义）
