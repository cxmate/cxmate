FROM golang:1.8

WORKDIR /go/src/github.com/cxmate/cxmate
COPY . .

RUN go-wrapper install

WORKDIR /cxmate

CMD ["cxmate"]
