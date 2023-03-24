# 第一阶段：使用 alpine 镜像作为构建环境
FROM golang:1.19-alpine as builder
WORKDIR /gofile

# 将依赖管理文件复制到容器中
COPY go.mod .

# 下载依赖
RUN go mod download

# 将项目代码复制到容器中
COPY . .

# 构建项目
RUN go build -ldflags "-s -w" -o goFile

# 第二阶段：使用最小的 alpine 镜像作为运行环境
FROM alpine:latest
WORKDIR /app

# 复制构建好的可执行文件到容器中
COPY --from=builder /gofile/goFile /app/
COPY --from=builder /gofile/templates /app/templates

# 暴露容器的 8089 端口
EXPOSE 8089

# 设置容器启动时要执行的命令
ENTRYPOINT [ "./goFile" ]
