FROM node:20.18-alpine3.20@sha256:d504f23acdda979406cf3bdbff0dff7933e5c4ec183dda404ed24286c6125e60 AS npm-builder

WORKDIR /build-dir

COPY frontend .

RUN npm install
RUN npm run build

FROM golang:1.23.2-alpine3.20@sha256:d21e934609de95ab75ba852128106ccf95ee7531e8b832b5f3b4e833d47a1ba2 AS go-builder

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