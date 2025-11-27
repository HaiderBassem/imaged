FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o imaged ./cmd/imaged-cli

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/imaged .
COPY --from=builder /app/configs/default.yaml ./config.yaml

VOLUME ["/data"]
WORKDIR /data

ENTRYPOINT ["./imaged"]