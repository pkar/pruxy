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

	var err error
	var p pruxy.Pruxy
	// choose between an etcd setup or os environment variables
	switch {
	case *etcdIPs != "":
		etcdHostList := strings.Split(*etcdIPs, ",")
		p, err = pruxy.NewEtcd(etcdHostList, *watchPrefix)
		if err != nil {
			log.Fatal(err)
		}
	default:
		p, err = pruxy.NewEnv(*watchPrefix)
		if err != nil {
			log.Fatal(err)
		}
	}

	converter := p.DefaultRequestConverter()
	proxy := pruxy.NewProxyWithRequestConverter(converter)

	// run a reverse-proxy server on http(s)://localhost:{port}/
	switch {
	case *certFile != "" && *keyFile != "":
		err = http.ListenAndServeTLS(":"+*port, *certFile, *keyFile, proxy)
		if err != nil {
			log.Fatal(err)
		}
	default:
		err = http.ListenAndServe(":"+*port, proxy)
		if err != nil {
			log.Fatal(err)
		}
	}
}
