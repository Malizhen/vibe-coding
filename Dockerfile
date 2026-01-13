# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go.mod
COPY go.mod ./

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -o blog .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates（用于 HTTPS）
RUN apk --no-cache add ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /app/blog .

# 复制模板和静态文件
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/content ./content
COPY --from=builder /app/data ./data

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV PORT=8080

# 运行
CMD ["./blog"]
