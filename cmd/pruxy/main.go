package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pkar/pruxy"
)

var (
	watchPrefix = flag.String("prefix", "pruxy", "where configs are (host->upstream)")
	port        = flag.String("port", "", "listen on port")
	certFile    = flag.String("certFile", "", "path to cert file")
	keyFile     = flag.String("keyFile", "", "path to key file")
	etcdIPs     = flag.String("etcd", "", "comma separated etcd ip address")
)

func redirectHttps(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, fmt.Sprintf("https://%s/%s", req.Host, req.URL.RequestURI()), http.StatusMovedPermanently)
}

func main() {
	flag.Parse()
	if *port == "" {
		log.Fatal("-port option required")
	}

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
		go func() {
			err := http.ListenAndServe(":80", http.HandlerFunc(redirectHttps))
			if err != nil {
				log.Fatal(err)
			}
		}()
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
