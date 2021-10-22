FROM golang:1.14
WORKDIR /go/src/github.com/diamnet/go

COPY . .
ENV GO111MODULE=on
RUN go install github.com/diamnet/go/tools/...
RUN go install github.com/diamnet/go/services/...
