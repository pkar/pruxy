package pruxy

import (
	"container/ring"
	"net/http"
	"os"
	"strings"
	"sync"

	log "github.com/pkar/pruxy/vendor/log"
)

// PruxyEnv holds meta on configurations for routing to upstream servers.
type PruxyEnv struct {
	Hosts       map[*HostPath]*ring.Ring
	mu          *sync.Mutex
	watchPrefix string
}

// NewEnv set-up environment variable based upstream hosts.
// Format should be
//   PREFIX_VAR="{Host}=upstream1:port1,upstream2:port2"
//   PRUXY_1="admin.dev.local=$127.0.0.1:8080,$127.0.0.1:8081" pruxy -prefix=PRUXY_
func NewEnv(prefix string) (*PruxyEnv, error) {
	p := &PruxyEnv{
		Hosts:       map[*HostPath]*ring.Ring{},
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
			u := strings.SplitN(tokens[0], "/", 2)
			host := u[0]
			var path string
			if len(u) > 1 {
				path = removeTrailingSlash("/" + u[1])
			} else {
				path = "/"
			}
			hostPath := &HostPath{host, path}

			upstreams := strings.Split(tokens[1], ",")

			p.Hosts[hostPath] = ring.New(len(upstreams))
			for _, upstream := range upstreams {
				p.Hosts[hostPath].Value = upstream
				p.Hosts[hostPath] = p.Hosts[hostPath].Next()
				log.Infof("added upstream %s%s -> %s", hostPath.Host, hostPath.Path, upstream)
			}
		}
	}

	return p, nil
}

// DefaultRequestConverter takes a request and converts it to an upstream
// one.
func (p *PruxyEnv) DefaultRequestConverter() func(*http.Request, *http.Request) {
	return func(originalRequest, proxy *http.Request) {
		p.convert(originalRequest, proxy)
	}
}

func (p *PruxyEnv) convert(originalRequest, proxy *http.Request) {
	originalHostPath := &HostPath{originalRequest.Host, originalRequest.URL.Path}

	p.mu.Lock()
	defer p.mu.Unlock()
	for hostPath, upstreams := range p.Hosts {
		if hostPath.Host == originalHostPath.Host && strings.HasPrefix(originalHostPath.Path, hostPath.Path) {
			upstreamHost := upstreams.Value.(string)
			upstreams = upstreams.Next()
			p.Hosts[hostPath] = upstreams
			proxy.URL.Host = upstreamHost
			proxy.URL.Path = strings.TrimPrefix(originalHostPath.Path, hostPath.Path)
		}
	}
}
