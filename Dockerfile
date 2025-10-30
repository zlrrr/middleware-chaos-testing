# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /mct cmd/mct/main.go

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -g 1000 mct && \
    adduser -D -u 1000 -G mct mct

WORKDIR /home/mct

# 从构建阶段复制二进制文件
COPY --from=builder /mct .
COPY --from=builder /app/configs ./configs

# 更改所有者
RUN chown -R mct:mct /home/mct

# 切换到非root用户
USER mct

# 设置入口点
ENTRYPOINT ["./mct"]

# 默认命令
CMD ["--help"]
