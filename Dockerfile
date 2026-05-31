FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gocachelab ./cmd/gocachelab

FROM alpine:3.22

RUN addgroup -S gocachelab && adduser -S -G gocachelab gocachelab
WORKDIR /app
COPY --from=build /out/gocachelab /usr/local/bin/gocachelab
RUN mkdir -p /app/data && chown -R gocachelab:gocachelab /app

USER gocachelab
EXPOSE 7379 8080
ENV GOCACHELAB_TCP_ADDR=0.0.0.0:7379
ENV GOCACHELAB_ADMIN_ADDR=0.0.0.0:8080
ENV GOCACHELAB_DATA_DIR=/app/data

ENTRYPOINT ["/usr/local/bin/gocachelab"]

