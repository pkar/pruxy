FROM golang

WORKDIR /go
ADD . src/github.com/pkar/pruxy
RUN go get ./...
RUN go install github.com/pkar/pruxy/cmd/pruxy

RUN rm -rf /go/src/*
CMD ["/go/bin/pruxy"]
