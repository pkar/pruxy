// Package pruxy is a simple reverse proxy that looks for
// changes in etcd to update it's configuration.
package pruxy

import ()

// Pruxy holds meta on configurations for routing to upstream servers.
type Pruxy interface {
	DefaultConverter() func(string) string
}
