FROM golang:1.11-alpine3.8 as builder

RUN set -x \
    && apk add --no-cache git \
    && mkdir -p /tmp

COPY . /qlrepl

WORKDIR /qlrepl

RUN set -x \
    && export CGO_ENABLED=0 \
    && go build -o /go/bin/qlrepl cmd/repl/main.go

# Executable image
FROM alpine:3.8

RUN apk add --update --no-cache \
    graphviz \
    ttf-freefont

COPY --from=builder /go/bin/qlrepl /usr/local/sbin/qlrepl

ADD examples /examples

ENTRYPOINT /usr/local/sbin/qlrepl