FROM node:25-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4 AS npm-builder

WORKDIR /build-dir

COPY frontend .

RUN npm ci
RUN npm run build

FROM golang:1.26.4-alpine@sha256:f23e8b227fb4493eabe03bede4d5a32d04092da71962f1fb79b5f7d1e6c2a17f AS go-builder

WORKDIR /app

COPY go.mod go.sum ./

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./rueckenwind ./cmd/rueckenwind

FROM gcr.io/distroless/static-debian13@sha256:3592aa8171c77482f62bbc4164e6a2d141c6122554ace66e5cc910cadb961ff0

WORKDIR /app

COPY --from=go-builder /app/rueckenwind ./rueckenwind

COPY --from=npm-builder /build-dir/dist ./frontend/dist

ENTRYPOINT [ "/app/rueckenwind" ]
