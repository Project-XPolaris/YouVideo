ARG GOLANG_VERSION=1.24-alpine
FROM golang:${GOLANG_VERSION} as builder

ARG GOPROXY=https://goproxy.cn
ENV GOPROXY=${GOPROXY}
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod go build -ldflags="-s -w" -o /youvideo ./main.go

FROM alpine:latest
RUN apk --no-cache add ffmpeg ca-certificates

WORKDIR /app
COPY --from=builder /youvideo .

ENTRYPOINT ["/app/youvideo", "run"]