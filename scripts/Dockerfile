# FROM golang:1.19.3-alpine as builder
FROM golang:alpine as builder

RUN set -eux \
    && mkdir -p /tmp/src \
    && cd /tmp/src \
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
    && git clone -b fiber --recursive --depth 1 https://github.com/bailangvvkruner/sonic . \
    # go get -u \
    # go get -u ./... \
    # 在构建阶段添加
    # go mod download \
    # && go list -m -u all \
    && go get -u ./... \
    # # 使用 govulncheck 检查漏洞并建议升级
    # && go install golang.org/x/vuln/cmd/govulncheck@latest \
    # && govulncheck ./... || true \
    # # 使用工具自动升级
    # && go install github.com/oligot/go-mod-upgrade@latest \
    # && go-mod-upgrade || true \
    # # 方法一：使用 -u=patch 升级补丁版本，-u 升级次要版本，但对主版本无效
    # go get -u github.com/disintegration/imaging
    # go get -u github.com/golang-jwt/jwt
    # 方法二：使用 @latest 但清除版本约束
    && go get github.com/golang-jwt/jwt/v5@latest \
    && go get github.com/disintegration/imaging@latest \
    && go mod tidy \
    # && CGO_ENABLED=0 GOOS=linux \
    && CGO_ENABLED=1 GOOS=linux \
        go build \
        -o $FILENAME \
        -ldflags="-s -w -extldflags -static" \
        # -ldflags="-s -w" \
        -trimpath . \
        && echo "Binary size after build:" \
        && du -b $FILENAME \
        && strip --strip-all $FILENAME \
        && echo "Binary size after stripping:" \
        && du -b $FILENAME \
        # && upx --best --lzma sonic \
        && (upx --best --lzma2 $FILENAME 2>/dev/null || echo "upx compression skipped") \
        && echo "Binary size after upx:" \
        && du -b $FILENAME \
    \
    && mkdir -p /app/conf \
    && mkdir /app/resources \
    && cp -r /tmp/srcsonic /app/ \
    && cp -r /tmp/srcconf /app/ \
    && cp -r /tmp/srcresources /app/ \
    && cp /tmp/srcscripts/docker_init.sh /app/


FROM alpine:latest as prod

COPY --from=builder /app/ /app/

RUN apk add --no-cache tzdata  ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

VOLUME /sonic
EXPOSE 8080

WORKDIR /sonic
CMD /app/docker_init.sh && /app/sonic -config /sonic/conf/config.yaml