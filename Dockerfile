ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /gomcp .

FROM debian:bookworm

RUN apt-get update && apt-get -y install ca-certificates

COPY --from=builder /gomcp /usr/local/bin/
CMD ["gomcp", "-api-addr", "0.0.0.0:8081", "sse"]
