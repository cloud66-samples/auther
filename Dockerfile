FROM golang:1.10

WORKDIR /go/src/auther
COPY main.go .
RUN go get -d -v ./...
RUN go build

CMD ["/go/src/auther/auther"]
