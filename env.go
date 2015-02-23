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
func NewEnv(prefix string) (*PruxyEnv, error) {
	p := &PruxyEnv{
		Hosts:       map[string]*ring.Ring{},
		mu:          &sync.Mutex{},
		watchPrefix: prefix,
	}

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if strings.HasPrefix(pair[0], prefix) {
			host := pair[0][len(prefix):]

			upstreams := strings.Split(pair[1], ",")
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
