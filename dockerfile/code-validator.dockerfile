FROM golang:1.19 as builder
WORKDIR /app
ADD . /app

RUN --mount=type=cache,target=/root/.cache/go-build go build -o bin/dispatcher-svc cmd/dispatcher/main.go && \
    go build -o bin/result-svc cmd/result/main.go && \
    go build -o bin/user-svc cmd/user/main.go

FROM ubuntu as dispatcher
WORKDIR /
COPY --from=builder /app/bin/dispatcher-svc /usr/local/bin
USER root

FROM ubuntu as result
WORKDIR /
COPY --from=builder /app/bin/result-svc /usr/local/bin
USER root

FROM ubuntu as user
WORKDIR /
COPY --from=builder /app/bin/user-svc /usr/local/bin
USER root