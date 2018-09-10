# build stage
FROM golang:1.10.3-alpine3.7 AS builder

RUN apk update && \
    apk add libressl-dev bash git ca-certificates curl

# establish a working directory
WORKDIR /go/src/github.com/deciphernow/gm-fabric-go/proxy-metrics-tester
ADD . /go/src/github.com/deciphernow/gm-fabric-go/proxy-metrics-tester

# vendor dependencies in the builder container
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    cd /go/src/github.com/deciphernow/gm-fabric-go/proxy-metrics-tester && \
    dep ensure --vendor-only -v

# build our binary
RUN go build -o proxy-metrics-tester .

# final stage
FROM alpine:3.7

RUN apk update && \
    apk add ca-certificates

WORKDIR /app

# Copy over the neccessary items from the builder container: binary, entrypoint script, and certs
COPY --from=builder /go/src/github.com/deciphernow/gm-fabric-go/proxy-metrics-tester/proxy-metrics-tester /app/

RUN addgroup -g 1000 -S metricstestergroup && \
    adduser -u 1000 -S metricstesteruser -G metricstestergroup && \
    chown -R metricstesteruser:metricstestergroup *

USER metricstesteruser

CMD ./proxy-metrics-tester
