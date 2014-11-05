// Package pruxy is a simple reverse proxy that looks for
// changes in etcd to update it's configuration.
package pruxy

import (
	"container/ring"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/go-etcd/etcd"
	log "github.com/golang/glog"
)

// Pruxy holds meta on configurations for routing to upstream servers.
type Pruxy struct {
	Hosts     map[string]*ring.Ring
	client    *etcd.Client
	mu        *sync.Mutex
	watchChan chan *etcd.Response
	stopChan  chan bool
	watchDir  string
}

// Set-up a connection to a etcd servers and initialize
// upstream hosts.
func NewPruxy(etcdHosts []string, dir string) (*Pruxy, error) {
	clientHosts := []string{}
	for _, host := range etcdHosts {
		clientHosts = append(clientHosts, fmt.Sprintf("http://%s", host))
	}

	p := &Pruxy{
		Hosts:     map[string]*ring.Ring{},
		watchChan: make(chan *etcd.Response),
		stopChan:  make(chan bool),
		mu:        &sync.Mutex{},
		watchDir:  dir,
	}

	p.client = etcd.NewClient(clientHosts)
	// create initial dir
	resp, err := p.client.CreateDir(p.watchDir, 0)
	log.Infof("adding key: %s resp: %v err: %s", p.watchDir, resp, err)

	// load in configuration on start
	err = p.load()
	if err != nil {
		log.Fatal(err, clientHosts)
	}
	// wait for changes
	go p.watch()
	return p, nil
}

// DefaultConverter returns a function which provides the
// host to upstream conversion.
func (p *Pruxy) DefaultConverter() func(string) string {
	return func(originalHost string) string {
		return p.convert(originalHost)
	}
}

func (p *Pruxy) convert(host string) string {
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

// watch waits for the configured dir in etcd for changes
// and reloads the configuration.
func (p *Pruxy) watch() {
	go p.client.Watch(p.watchDir, 0, true, p.watchChan, p.stopChan)
	log.Infof("watching %s", p.watchDir)
	for {
		r := <-p.watchChan
		if r == nil {
			log.Info("no change")
			continue
		}
		err := p.load()
		if err != nil {
			log.Error(err)
		}
	}
}

// load reloads the proxy config from etcd
func (p *Pruxy) load() error {
	response, err := p.client.Get(p.watchDir, false, true)
	if err != nil {
		log.Error(err)
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
			log.V(2).Infof("%+v", upstreamNode)
			upstream := strings.Split(upstreamNode.Key, "/")[3]
			if upstream != "" {
				p.Hosts[host].Value = upstream
				p.Hosts[host] = p.Hosts[host].Next()
				log.Infof("added upstream %s -> %s", host, upstream)
			}
		}
	}
	return nil
}
