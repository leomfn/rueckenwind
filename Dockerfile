FROM golang:1.23.1-alpine AS builder

RUN apk add build-base

WORKDIR /app

COPY go.mod go.sum ./

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./windeows ./cmd/rueckenwind

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/windeows ./windeows

COPY static/ ./static
COPY templates/ ./templates

ENTRYPOINT [ "/app/windeows" ]