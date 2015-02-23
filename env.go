package pruxy

import (
	"container/ring"
	"os"
	"strings"
	"sync"

	log "github.com/pkar/pruxy/vendor/log"
)

// Pruxy holds meta on configurations for routing to upstream servers.
type PruxyEnv struct {
	Hosts       map[string]*ring.Ring
	mu          *sync.Mutex
	watchPrefix string
}

// Set-up environment variable based
// upstream hosts.
// Format should be
// PREFIX_VAR="{Host}=upstream1:port1,upstream2:port2"
// PRUXY_1="admin.dev.local=$127.0.0.1:8080,$127.0.0.1:8081" pruxy -prefix=PRUXY_
func NewEnv(prefix string) (*PruxyEnv, error) {
	p := &PruxyEnv{
		Hosts:       map[string]*ring.Ring{},
		mu:          &sync.Mutex{},
		watchPrefix: prefix,
	}

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if strings.HasPrefix(pair[0], prefix) {
			tokens := strings.Split(pair[1], "=")
			if len(tokens) != 2 {
				log.Error("invalid host ip format ", env)
				continue
			}

			host := tokens[0]
			upstreams := strings.Split(tokens[1], ",")

			p.Hosts[host] = ring.New(len(upstreams))
			for _, upstream := range upstreams {
				p.Hosts[host].Value = upstream
				p.Hosts[host] = p.Hosts[host].Next()
				log.Infof("added upstream %s -> %s", host, upstream)
			}
		}
	}

	return p, nil
}

// DefaultConverter returns a function which provides the
// host to upstream conversion.
func (p *PruxyEnv) DefaultConverter() func(string) string {
	return func(originalHost string) string {
		return p.convert(originalHost)
	}
}

func (p *PruxyEnv) convert(host string) string {
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
