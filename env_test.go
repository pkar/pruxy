package pruxy

import (
	"net/http"
	"os"
	"testing"
)

func TestNewEnv(t *testing.T) {
	os.Setenv("TEST_1", "abc.com=127.0.0.1:8081,127.0.0.8082")
	os.Setenv("TEST_2", "abc.com/abc/123=127.0.0.1,127.0.0.1:8080")
	os.Setenv("TEST_3", "abc.com/abc/123")
	defer os.Unsetenv("TEST_1")
	defer os.Unsetenv("TEST_2")
	defer os.Unsetenv("TEST_3")

	p, err := NewEnv("TEST_")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Hosts) != 2 {
		t.Fatalf("p.Hosts not initialized correctly, should have 2 hosts got %d", len(p.Hosts))
	}
	for _, hostPath := range p.Hosts {
		if hostPath.Len() != 2 {
			t.Fatalf("p.Hosts upstreams not initialized correctly, should have 2 got %d", hostPath.Len())
		}
	}
}

func TestEnvDefaultRequestConverter(t *testing.T) {
	os.Setenv("TEST_1", "abc.com=127.0.0.1:8081,127.0.0.8082")
	os.Setenv("TEST_2", "abc.com/abc/123=127.0.0.1")

	p, err := NewEnv("TEST_")
	if err != nil {
		t.Fatal(err)
	}
	p.DefaultRequestConverter()
}

func TestEnvConvert(t *testing.T) {
	os.Setenv("TEST_1", "abc.com=127.0.0.1:8081,127.0.0.1:8082")
	os.Setenv("TEST_2", "abc.com/abc/123=127.0.0.1")

	p, _ := NewEnv("TEST_")
	convertFunc := p.DefaultRequestConverter()
	in, _ := http.NewRequest("GET", "http://abc.com/", nil)
	out := copyRequest(in)
	convertFunc(in, out)
	if out.URL.String() != "http://127.0.0.1:8081" {
		t.Fatalf("expected http://127.0.0.1:8081, got %s", out.URL.String())
	}
}
