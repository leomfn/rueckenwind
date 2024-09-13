FROM golang:1.23.1-alpine AS builder

RUN apk add build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/mattn/go-sqlite3

COPY *.go ./

RUN go build -ldflags="-extldflags=-static" -v -o ./windeows

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/windeows ./windeows

COPY *.html ./

ENTRYPOINT [ "/app/windeows" ]