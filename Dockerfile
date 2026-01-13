# 通用Dockerfile，用于构建所有服务
# 使用方式: docker build -f Dockerfile --build-arg SERVICE=user_service -t user-service .

FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建参数：指定要构建的服务
ARG SERVICE

# 构建服务
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/${SERVICE} ./${SERVICE}/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/${SERVICE} /root/${SERVICE}

# 暴露端口（通过环境变量配置）
EXPOSE 8080 8081 8082 8083

# 运行服务
CMD ["./${SERVICE}"]

