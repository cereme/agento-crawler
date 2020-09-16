FROM golang:1.14-alpine as builder
MAINTAINER ceremebsy@gmail.com
RUN apk update && apk add git && apk add ca-certificates

WORKDIR /usr/src/app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/src/app/main /main

CMD [ "/main" ]