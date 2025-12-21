FROM node:25-alpine@sha256:405485d01af4f7087dcf0029fcc3444a260585f6900cff1024f555d0d64bf756 AS npm-builder

WORKDIR /build-dir

COPY frontend .

RUN npm install
RUN npm run build

FROM golang:1.25.5-alpine3.20@sha256:f8f784f478e37b032640fcd0fa31b1f1bd0d79dd36a746161b62c52e8763d290 AS go-builder

RUN apk add build-base

WORKDIR /app

COPY go.mod go.sum ./

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./rueckenwind ./cmd/rueckenwind

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=go-builder /app/rueckenwind ./rueckenwind

COPY --from=npm-builder /build-dir/dist ./frontend/dist

ENTRYPOINT [ "/app/rueckenwind" ]
