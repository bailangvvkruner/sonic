# FROM golang:1.19.3-alpine as builder
FROM golang:alpine as builder

RUN apk --no-cache add \
    git \
    ca-certificates \
    gcc \
    g++ \
    # 包含strip命令
    binutils \
    upx

# COPY . /go/src/github.com/go-sonic/sonic/
WORKDIR /go/src/github.com/go-sonic/sonic

RUN \
    git clone -b master --recursive --depth 1 https://github.com/bailangvvkruner/sonic . \
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
    && CGO_ENABLED=1 GOOS=linux \
        go build \
        -o sonic \
        # -ldflags="-s -w" \
        -ldflags="-s -w -extldflags -static" \
        -trimpath . \
        && echo "Binary size after build:" \
        && du -b sonic \
        && strip --strip-all sonic \
        && echo "Binary size after stripping:" \
        && du -b sonic \
        && upx --best --lzma sonic \
        && echo "Binary size after upx:" \
        && du -b sonic

RUN mkdir -p /app/conf && \
    mkdir /app/resources && \
    cp -r /go/src/github.com/go-sonic/sonic/sonic /app/ && \
    cp -r /go/src/github.com/go-sonic/sonic/conf /app/ && \
    cp -r /go/src/github.com/go-sonic/sonic/resources /app/ && \
    cp /go/src/github.com/go-sonic/sonic/scripts/docker_init.sh /app/


FROM alpine:latest as prod

COPY --from=builder /app/ /app/

RUN apk add --no-cache tzdata  ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

VOLUME /sonic
EXPOSE 8080

WORKDIR /sonic
CMD /app/docker_init.sh && /app/sonic -config /sonic/conf/config.yaml