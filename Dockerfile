FROM golang:1.23.5-alpine3.21  as build

LABEL maintainer="longfei6671@163.com"

RUN apk add  --update-cache  libc-dev git gcc musl-dev sqlite-dev

WORKDIR /go/src/app

ADD . /go/src/app/DouYinBot/

RUN cd DouYinBot && go mod download && go build -ldflags="-s -w" -o douyinbot main.go

FROM alpine:3.21

LABEL maintainer="longfei6671@163.com"

COPY --from=build /go/src/app/DouYinBot/douyinbot /var/www/douyinbot/

RUN apk add --no-cache sqlite-libs
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN chmod +x /var/www/douyinbot/douyinbot

WORKDIR /var/www/douyinbot/

EXPOSE 9080

CMD ["/var/www/douyinbot/douyinbot","--config-file","/var/www/douyinbot/conf/app.conf","--data-file","/var/www/douyinbot/data/douyinbot.db"]

