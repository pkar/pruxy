package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/pkar/pruxy"
	log "github.com/pkar/pruxy/vendor/log"
)

var (
	watchPrefix = flag.String("prefix", "pruxy", "where configs are (host->upstream)")
	port        = flag.String("port", "6000", "listen on port")
	certFile    = flag.String("certFile", "", "path to cert file")
	keyFile     = flag.String("keyFile", "", "path to key file")
	etcdIPs     = flag.String("etcd", "", "comma separated etcd ip address")
)

func main() {
	flag.Parse()
	var p pruxy.Pruxy
	var err error

	if *etcdIPs != "" {
		etcdHostList := strings.Split(*etcdIPs, ",")
		p, err = pruxy.NewEtcd(etcdHostList, *watchPrefix)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		p, err = pruxy.NewEnv(*watchPrefix)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Returns one of registered upstream hosts
	hostConverter := p.DefaultConverter()

	proxy := pruxy.NewProxyWithHostConverter(hostConverter)

	// Runs a reverse-proxy server on http(s)://localhost:{port}/
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
