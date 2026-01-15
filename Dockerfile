# 多阶段构建 Dockerfile for Prediction Relayer Service
# ⚠️ 注意：此 Dockerfile 需要从项目根目录构建
# 构建命令：docker build -f prediction-relayer-service/Dockerfile -t image:tag .
# Stage 1: 构建阶段
FROM golang:1.25-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git make protobuf protobuf-dev

# 设置工作目录
WORKDIR /workspace/prediction-relayer-service

# 复制服务的 go mod 文件
COPY prediction-relayer-service/go.mod prediction-relayer-service/go.sum ./

# 下载依赖
RUN go mod download

# 复制服务源代码
COPY prediction-relayer-service/ .

# 更新 go.mod（确保依赖关系正确）
RUN go mod tidy

# 生成 proto 和 wire 代码（如果需要）
RUN make api wire || true

# 构建二进制文件（包含 wire_gen.go 如果存在）
RUN if [ -f cmd/server/wire_gen.go ]; then \
      CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go cmd/server/wire_gen.go; \
    else \
      CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go; \
    fi

# Stage 2: 运行阶段
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /workspace/prediction-relayer-service/server .
COPY --from=builder /workspace/prediction-relayer-service/configs ./configs

# 创建日志目录
RUN mkdir -p logs && chown -R app:app /app

# 切换到非 root 用户
USER app

# 暴露端口（根据实际配置调整）
EXPOSE 8100 9100

# 启动服务
CMD ["./server", "-conf", "configs/config.yaml"]

