# 使用 golang 官方镜像提供 Go 运行环境，并且命名为 buidler 以便后续引用
FROM golang:1.18-alpine as builder

# 安装 git
RUN apk --no-cache add git

# 设置工作目录
WORKDIR /app

# 将当前项目所在目录代码拷贝到镜像中
COPY . .

# 下载依赖
RUN go mod download

# 构建二进制文件，添加来一些额外参数以便可以在 Alpine 中运行它
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go_api


# 下面是第二阶段的镜像构建，和之前保持一致
FROM alpine:latest

# 安装相关软件
RUN apk update && apk add --no-cache bash supervisor ca-certificates

# 和上个阶段一样设置工作目录
RUN mkdir /app
WORKDIR /app

# 而是从上一个阶段构建的 builder容器中拉取
COPY --from=builder /app/go_api .
ADD supervisord.conf /etc/supervisord.conf
ADD config.dev.json /app/config.dev.json
RUN cp config.dev.json config.json


# 通过 Supervisor 管理服务
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
