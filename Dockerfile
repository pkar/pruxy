FROM golang

WORKDIR /go
ADD . src/github.com/pkar/pruxy
RUN go get ./...

RUN mv /go/src/github.com/pkar/pruxy/certs /go/
RUN rm -rf /go/src/*
CMD ["/go/bin/pruxy"]
