# 使用官方Go镜像作为构建阶段
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mtRSSConverter .

# 使用alpine作为运行阶段基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/mtRSSConverter .

# 暴露应用程序使用的端口
EXPOSE 8080

# 设置启动命令
CMD ["./mtRSSConverter"]