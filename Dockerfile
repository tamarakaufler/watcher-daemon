FROM golang:1.15.8-alpine AS builder

ENV VERSION=unknown
ENV GIT_SHA=unknown
ENV NAME=watcher-daemon
ENV TIMESTAMP=unknown
ENV LD_FLAGS="-w -s -X main.Timestamp=${TIMESTAMP} -X main.Version=${VERSION}  -X main.GitSHA=${GIT_SHA} -X main.ServiceName=${NAME}"

RUN echo ${LD_FLAGS}

RUN apk add --no-cache tzdata git make

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux

COPY . .
RUN go build -o ./bin/watcher-daemon ./cmd/watcher-daemon/main.go


FROM alpine:3.13
WORKDIR /app/basepath
COPY --from=builder /app/bin/watcher-daemon /bin/watcher-daemon
ENTRYPOINT [ "/bin/watcher-daemon" ]
