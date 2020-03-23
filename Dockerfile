#指定基础镜像
FROM golang:1.13.1 as builder

#工作目录
WORKDIR $GOPATH/src/minestat
COPY . $GOPATH/src/minestat

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go get && go build -a -ldflags '-extldflags "-static"' -o /minestat .

FROM alpine

WORKDIR /opt/app/

COPY --from=builder  /minestat /opt/app/minestat

RUN chmod -R 755 /opt/app/minestat && \
    mkdir /etc/minestat

# 设置时区
RUN apk update && apk add ca-certificates && \
    apk add tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

VOLUME /root/.minestat

ENTRYPOINT ["./minestat"]