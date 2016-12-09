FROM golang:1.7

RUN mkdir -p /go/src/github.com/eventials/go-tus

WORKDIR /go/src/github.com/eventials/go-tus

RUN go get github.com/stretchr/testify
RUN go get github.com/tus/tusd
