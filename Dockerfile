
FROM golang:1.11-alpine AS build_base

RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /go/src/github.com/ragurney/term-check

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_base AS server_builder

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install -a -tags netgo -ldflags '-w -extldflags "-static"' ./cmd/term-check

FROM alpine AS term-check

RUN apk add ca-certificates

COPY --from=server_builder /go/bin/term-check /bin/term-check
COPY --from=server_builder /go/src/github.com/ragurney/term-check/config.yaml .

ENTRYPOINT ["/bin/term-check", "--config", "config.yaml"]
