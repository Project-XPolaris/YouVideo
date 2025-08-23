ARG GOLANG_VERSION=1.22-alpine
FROM docker.io/library/golang:${GOLANG_VERSION} as builder

ARG GOPROXY=https://goproxy.cn
ENV GOPROXY=${GOPROXY}
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod go build -ldflags="-s -w" -o /youvideo ./main.go

FROM docker.io/library/alpine:3.19
RUN apk --no-cache add ffmpeg ca-certificates

WORKDIR /app
COPY --from=builder /youvideo .

ENTRYPOINT ["/app/youvideo", "run"]