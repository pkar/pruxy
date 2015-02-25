package pruxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestNewProxyWithHostConverter(t *testing.T) {
	NewProxyWithHostConverter(func(a string) string {
		return a
	})
}

func TestNewProxyWithRequestConverter(t *testing.T) {
	os.Setenv("TEST_1", "abc.com=127.0.0.1:8081,127.0.0.8082")
	defer os.Unsetenv("TEST_1")

	var proxyTests = []struct {
		method string
		host   string
		path   string
	}{
		{"GET", "abc.com", "/"},
	}
	p, _ := NewEnv("TEST_")
	converter := p.DefaultRequestConverter()
	proxy := NewProxyWithRequestConverter(converter)

	for _, tt := range proxyTests {
		r := &http.Request{
			Method: tt.method,
			Host:   tt.host,
			URL: &url.URL{
				Path: tt.path,
			},
		}
		rr := httptest.NewRecorder()
		proxy.ServeHTTP(rr, r)
	}
}
