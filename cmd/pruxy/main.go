package main

import (
	"flag"
	"net/http"
	"strings"

	log "github.com/golang/glog"
	"github.com/pkar/pruxy"
)

var (
	watchDir = flag.String("dir", "/pruxy", "where configs are (host->upstream)")
	port     = flag.String("port", "6000", "listen on port")
	etcdIPs  = flag.String("etcd", "172.17.42.1:4001", "comma separated etcd ip address default 172.17.42.1")
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	etcdHostList := strings.Split(*etcdIPs, ",")
	p, err := pruxy.NewPruxy(etcdHostList, *watchDir)
	if err != nil {
		log.Fatal(err)
	}

	// Returns one of registered upstream hosts
	hostConverter := p.DefaultConverter()

	// Runs a reverse-proxy server on http://localhost:{port}/
	proxy := pruxy.NewProxyWithHostConverter(hostConverter)
	err = http.ListenAndServe(":"+*port, proxy)
	if err != nil {
		log.Fatal(err)
	}
}
