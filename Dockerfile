
FROM golang

ENV ETCD 172.17.42.1:4001

ADD . /go/src/github.com/pkar/pruxy
RUN cd /go && go get ./...
RUN go install github.com/pkar/pruxy/cmd/pruxy

RUN rm -rf /go/src/*
CMD ["/go/bin/pruxy"]
