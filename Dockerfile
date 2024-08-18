FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download

RUN go build -o ./windeows

FROM gcr.io/distroless/static

WORKDIR /app

COPY *.html ./
COPY --from=builder /app/windeows .

ENTRYPOINT [ "/app/windeows" ]