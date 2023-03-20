FROM golang:1.19 as builder
WORKDIR /app
ADD . /app
RUN --mount=type=cache,target=/root/.cache/go-build go build -tags=javascript -o bin/sandbox cmd/sandbox-test/main.go && \
    go build -tags=javascript -o bin/validator cmd/validator-test/main.go && \
    go build -tags=javascript -o bin/code-match cmd/code-match/main.go

FROM node
WORKDIR /app
# git clone https://github.com/ioi/isolate.git && make install -C isolate && rm -rf isolate \
# RUN echo "deb http://mirrors.aliyun.com/debian bullseye main" > /etc/apt/sources.list &&\
#     echo "deb http://mirrors.aliyun.com/debian bullseye-updates main" >> /etc/apt/sources.list

RUN apt-get update && apt-get install -y libcap-dev && apt-get clean && \
    curl -L -o isolate.zip https://github.91chi.fun//https://github.com/ioi/isolate/archive/refs/heads/master.zip && \
    unzip isolate.zip && \
    make install -C isolate-master && \
    rm -rf isolate-master isolate.zip

COPY --from=builder /app/bin/* /usr/local/bin
USER root