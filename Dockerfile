# ==========================================
# 阶段 1：构建静态二进制文件（支持多架构快速交叉编译）
# ==========================================
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# 复制依赖定义和源码
COPY go.mod ./
COPY *.go ./

# 交叉编译极简静态二进制文件
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o notify-service .

# ==========================================
# 阶段 2：运行环境（轻量级 Alpine）
# ==========================================
FROM alpine:latest

# 安装 HTTPS 根证书与时区支持
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段拷贝编译好的二进制
COPY --from=builder /build/notify-service /app/notify-service

# 暴露端口
EXPOSE 5755

# 启动服务
ENTRYPOINT ["/app/notify-service"]
