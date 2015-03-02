package pruxy

import (
	"net/http"
	"net/http/httptest"
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
		url    string
	}{
		{"GET", "http://abc.com"},
		{"GET", "127.0.0.1:8000"},
	}
	p, _ := NewEnv("TEST_")
	converter := p.DefaultRequestConverter()
	proxy := NewProxyWithRequestConverter(converter)

	for _, tt := range proxyTests {
		r, err := http.NewRequest(tt.method, tt.url, nil)
		if err != nil {
			t.Error(err)
		}
		r.RemoteAddr = "abc:8000"
		r.Header.Add("X-Forwarded-For", "129.78.138.66")
		rr := httptest.NewRecorder()
		proxy.ServeHTTP(rr, r)
	}
}
