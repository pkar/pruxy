package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/pkar/pruxy"
	log "github.com/pruxy/vendor/log"
)

var (
	watchDir = flag.String("dir", "/pruxy", "where configs are (host->upstream)")
	port     = flag.String("port", "6000", "listen on port")
	certFile = flag.String("certFile", "", "path to cert file")
	keyFile  = flag.String("keyFile", "", "path to key file")
	etcdIPs  = flag.String("etcd", "", "comma separated etcd ip address")
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
	if *certFile != "" && *keyFile != "" {
		err = http.ListenAndServeTLS(":"+*port, *certFile, *keyFile, proxy)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err = http.ListenAndServe(":"+*port, proxy)
	if err != nil {
		log.Fatal(err)
	}
}
