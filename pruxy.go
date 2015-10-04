// Package pruxy is a simple reverse proxy that looks for
// changes in etcd to update it's configuration.
package pruxy

import (
	"log"
	"net/http"
	"strings"
)

// Pruxy holds meta on configurations for routing to upstream servers.
type Pruxy interface {
	DefaultRequestConverter() func(*http.Request, *http.Request)
}

// HostPath is used as a key in Pruxy.Hosts for routing.
type HostPath struct {
	Host string
	Path string
}

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

// removeTrailingSlash removes a slash at the end of paths
// only if the path is longer than 1. For instance / would
// remain / but /a/ would become /a
func removeTrailingSlash(path string) string {
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}
