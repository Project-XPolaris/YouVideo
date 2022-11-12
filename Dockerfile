ARG GOLANG_VERSION=1.18
FROM golang:${GOLANG_VERSION}-buster as builder
ARG GOPROXY=https://goproxy.cn
WORKDIR ${GOPATH}/src/github.com/projectxpolaris/youvideo

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ${GOPATH}/bin/youvideo ./main.go

FROM ubuntu

COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

RUN apt update && apt install -y ffmpeg
COPY --from=builder /go/bin/youvideo /usr/local/bin/youvideo

ENTRYPOINT ["/usr/local/bin/youvideo","run"]