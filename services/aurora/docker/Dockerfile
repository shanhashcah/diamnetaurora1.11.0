FROM golang:1.14-stretch AS builder

WORKDIR /go/src/github.com/diamnet/go
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go install github.com/diamnet/go/services/aurora

FROM ubuntu:18.04

ENV DIAMNET_CORE_VERSION 14.1.1-1355-5b2989d3
ENV DIAMNET_CORE_BINARY_PATH /usr/local/bin/diamnet-core

# ca-certificates are required to make tls connections
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget


RUN apt-get install -y --no-install-recommends libpqxx-4.0v5 curl
RUN wget -O diamnet-core.deb https://s3.amazonaws.com/diamnet.org/releases/diamnet-core/diamnet-core-${DIAMNET_CORE_VERSION}_amd64.deb
RUN dpkg -i diamnet-core.deb
RUN rm diamnet-core.deb

RUN apt-get clean

COPY --from=builder /go/bin/aurora ./

ENTRYPOINT ["./aurora"]
