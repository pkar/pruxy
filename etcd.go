package pruxy

import (
	"container/ring"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-etcd/etcd"
)

// Pruxy holds meta on configurations for routing to upstream servers.
type PruxyEtcd struct {
	Hosts       map[string]*ring.Ring
	client      *etcd.Client
	mu          *sync.Mutex
	watchPrefix string
}

// Set-up a connection to a etcd servers and initialize
// upstream hosts.
func NewEtcd(etcdHosts []string, prefix string) (*PruxyEtcd, error) {
	clientHosts := []string{}
	for _, host := range etcdHosts {
		clientHosts = append(clientHosts, fmt.Sprintf("http://%s", host))
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = fmt.Sprintf("/%s", prefix)
	}

	p := &PruxyEtcd{
		Hosts:       map[string]*ring.Ring{},
		mu:          &sync.Mutex{},
		watchPrefix: prefix,
	}

	p.client = etcd.NewClient(clientHosts)
	// create initial prefix, ignore err if already defined
	resp, err := p.client.CreateDir(p.watchPrefix, 0)
	log.Printf("adding key: %s resp: %v err: %v", p.watchPrefix, resp, err)

	// load in configuration on start
	err = p.load()
	if err != nil {
		log.Println("err:", err)
		return nil, err
	}
	// wait for changes
	go p.watch()
	return p, nil
}

// DefaultRequestConverter
func (p *PruxyEtcd) DefaultRequestConverter() func(*http.Request, *http.Request) {
	return func(originalRequest, proxy *http.Request) {
		p.convert(originalRequest, proxy)
	}
}

func (p *PruxyEtcd) convert(originalRequest, proxy *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if upstreams, ok := p.Hosts[originalRequest.Host]; ok {
		upstreamHost := upstreams.Value.(string)
		upstreams = upstreams.Next()
		p.Hosts[originalRequest.Host] = upstreams
		proxy.URL.Host = upstreamHost
	}
}

func (p *PruxyEtcd) convertHost(host string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if upstreams, ok := p.Hosts[host]; ok {
		upstreamHost := upstreams.Value.(string)
		upstreams = upstreams.Next()
		p.Hosts[host] = upstreams
		return upstreamHost
	}
	return ""
}

// watch waits for the configured prefix in etcd for changes
// and reloads the configuration.
func (p *PruxyEtcd) watch() {
	watchChan := make(chan *etcd.Response)
	stopChan := make(chan bool)
	go func() {
		for {
			_, err := p.client.Watch(p.watchPrefix, 0, true, watchChan, stopChan)
			if err != nil {
				log.Println(err)
			}
			watchChan = make(chan *etcd.Response)
			stopChan = make(chan bool)
		}
	}()
	log.Printf("watching %s", p.watchPrefix)
	nErrs := 0
	for {
		select {
		case resp := <-watchChan:
			if resp != nil {
				err := p.load()
				if err != nil {
					log.Println("err:", err)
				}
			}
			nErrs++
			// nil here seems to mean etcd not found
			if nErrs > 10 {
				return
			}
		case <-stopChan:
			log.Println("stop watching")
			return
		}
	}
}

// load reloads the proxy config from etcd
func (p *PruxyEtcd) load() error {
	response, err := p.client.Get(p.watchPrefix, false, true)
	if err != nil {
		log.Println("err:", err)
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// reinit host configs
	p.Hosts = map[string]*ring.Ring{}
	for _, hostNode := range response.Node.Nodes {
		host := strings.Split(hostNode.Key, "/")[2]
		p.Hosts[host] = ring.New(len(hostNode.Nodes))
		for _, upstreamNode := range hostNode.Nodes {
			upstream := strings.Split(upstreamNode.Key, "/")[3]
			if upstream != "" {
				p.Hosts[host].Value = upstream
				p.Hosts[host] = p.Hosts[host].Next()
				log.Printf("added upstream %s -> %s", host, upstream)
			}
		}
	}
	return nil
}
