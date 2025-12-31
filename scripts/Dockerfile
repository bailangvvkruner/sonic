# FROM golang:1.19.3-alpine as builder
FROM golang:alpine as builder

RUN set -eux \
    && mkdir -p /go/src/github.com/go-sonic/sonic/ \
    && cd /go/src/github.com/go-sonic/sonic/ \
    && FILENAME=sonic \
    && apk add --no-cache --no-scripts --virtual .build-deps \
    git \
    ca-certificates \
    build-base \
    # 包含strip命令
    binutils \
    # upx
    \
    # 尝试安装 upx，如果不可用则继续（某些架构可能不支持）
    && apk add --no-cache --no-scripts --virtual .upx-deps \
        upx 2>/dev/null || echo "upx not available, skipping compression" \
    \
    # # COPY . /go/src/github.com/go-sonic/sonic/
    # WORKDIR /go/src/github.com/go-sonic/sonic
    \
    # && set -eux \
    && git clone -b strip --recursive --depth 1 https://github.com/bailangvvking/sonic . \
    # 切换到sonic_fiber_v2目录进行构建
    && cd sonic_fiber_v2 \
    && go get -u ./... \
    && go get github.com/golang-jwt/jwt/v5@latest \
    && go get github.com/disintegration/imaging@latest \
    && go mod tidy \
    && CGO_ENABLED=1 GOOS=linux \
        go build \
        -o $FILENAME \
        -ldflags="-s -w -extldflags -static" \
        -trimpath . \
        && echo "Binary size after build:" \
        && du -b $FILENAME \
        && strip --strip-all $FILENAME \
        && echo "Binary size after stripping:" \
        && du -b $FILENAME \
        && (upx --best --lzma $FILENAME 2>/dev/null || echo "upx compression skipped") \
        && echo "Binary size after upx:" \
        && du -b $FILENAME \
    \
    && mkdir -p /app/conf \
    && mkdir /app/resources \
    && cp -r $FILENAME /app/ \
    && cp -r conf/* /app/conf/ \
    && cp -r resources/* /app/resources/ \
    && cp /go/src/github.com/go-sonic/sonic/scripts/docker_init.sh /app/


FROM alpine:latest as prod

COPY --from=builder /app/ /app/

RUN apk add --no-cache tzdata  ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

VOLUME /sonic
EXPOSE 8080

WORKDIR /app
# CMD ./docker_init.sh && ./sonic -config /sonic/conf/config.yaml
CMD ./docker_init.sh && find / -name "sonic" -type f -executable -exec {} -config /sonic/conf/config.yaml

