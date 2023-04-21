FROM golang:1.19 as builder
WORKDIR /app
ADD . /app
RUN --mount=type=cache,target=/root/.cache/go-build go build -tags=python -o bin/actuator cmd/actuator/main.go && \
    go build -tags=python -o bin/code-match cmd/code-match/main.go

FROM golang:1.19
WORKDIR /app

RUN apt-get update && apt-get install -y libcap-dev && apt-get clean && \
    curl -L -o isolate.zip https://github.com/ioi/isolate/archive/refs/heads/master.zip && \
    unzip isolate.zip && \
    make install -C isolate-master && \
    rm -rf isolate-master isolate.zip

COPY --from=builder /app/bin/* /usr/local/bin
USER root