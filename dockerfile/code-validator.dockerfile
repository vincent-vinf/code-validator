FROM golang:1.19 as builder

WORKDIR /app

ADD . /app

RUN mkdir -p "bin" && \
    go build -o bin/access-service cmd/access/main.go && \
    go build -o bin/spike-service cmd/spike/main.go && \
    go build -o bin/user-service cmd/user/main.go && \
    go build -o bin/admin-service cmd/admin/main.go && \
    go build -o bin/order-service cmd/order/main.go

# gcr.io/distroless/static
# access
FROM ubuntu as access
WORKDIR /
COPY --from=builder /app/bin/access-service /
USER root
