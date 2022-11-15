FROM golang:1.18 as builder
WORKDIR /app
ADD . /app
RUN --mount=type=cache,target=/root/.cache/go-build go build -o bin/sandbox cmd/sandbox-test/main.go && \
    go build -o bin/validator cmd/validator-test/main.go && \
    go build -o bin/code-validator cmd/code-validator/main.go && \
    go build -o bin/pipeline cmd/pipeline-test/main.go

FROM node
WORKDIR /app
#RUN apt-get update && apt-get install -y libcap-dev && rm -rf /var/lib/apt/lists/* \
#    && git clone https://github.com/ioi/isolate.git && make install -C isolate && rm -rf isolate \
RUN apt-get update && apt-get install -y libcap-dev && rm -rf /var/lib/apt/lists/* && \
    curl -L -o isolate.zip https://github.91chi.fun//https://github.com/ioi/isolate/archive/refs/heads/master.zip && \
    unzip isolate.zip && \
    make install -C isolate-master && \
    rm -rf isolate-master isolate.zip

COPY --from=builder /app/bin/* /usr/local/bin
USER root