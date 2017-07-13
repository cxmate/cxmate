FROM golang:1.8

WORKDIR /go/src/cxmate
COPY . .

RUN go-wrapper download
RUN go-wrapper install

WORKDIR /cxmate

CMD ["cxmate"]
