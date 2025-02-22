ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
COPY --from=builder /go/bin/goose /usr/local/bin/
COPY migrations /migrations

CMD ["run-app"]
