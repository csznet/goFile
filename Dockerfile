FROM golang:1.19
WORKDIR /goFile
# 将依赖管理文件复制到容器中
COPY go.mod .
COPY go.sum .
# 下载依赖
RUN go mod download
# 将项目代码复制到容器中
COPY . .
# 构建项目
RUN go build -o app
# 设置容器启动时要执行的命令
CMD ["/goFile/goFile"]