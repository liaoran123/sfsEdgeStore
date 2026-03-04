# 使用官方Go镜像作为构建环境
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o sfsdb-edgex-adapter .

# 使用Alpine作为基础镜像
FROM alpine:latest

# 安装必要的依赖
RUN apk add --no-cache ca-certificates

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/sfsdb-edgex-adapter /app/

# 创建数据目录
RUN mkdir -p /app/edgex_data

# 暴露健康检查端口
EXPOSE 8081

# 设置环境变量
ENV EDGEX_DB_PATH=/app/edgex_data
ENV EDGEX_MQTT_BROKER=tcp://localhost:1883
ENV EDGEX_MQTT_TOPIC=edgex/events/core/#
ENV EDGEX_CLIENT_ID=sfsdb-edgex-adapter

# 运行应用
CMD ["./sfsdb-edgex-adapter"]