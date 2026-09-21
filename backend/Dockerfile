# 阶段一：使用完整的 Go 环境编译程序
FROM golang:1.26-alpine AS builder

WORKDIR /src

# 先复制依赖文件。业务代码变化时，可以继续使用依赖缓存。
COPY go.mod go.sum ./
RUN go mod download

# 复制源码，编译出 Linux 可执行文件。
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=1 go build -p=1 -trimpath -ldflags="-s -w" -o /out/video-feedsystem .


# 阶段二：只保留运行程序需要的内容
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app \
    && mkdir -p /app/data \
    && chown -R app:app /app

WORKDIR /app

ENV TZ=Asia/Shanghai

COPY --from=builder --chown=app:app /out/video-feedsystem ./video-feedsystem

USER app

EXPOSE 8080

ENTRYPOINT ["./video-feedsystem"]
