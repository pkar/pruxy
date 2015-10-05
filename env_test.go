package pruxy

import (
	"net/http"
	"os"
	"testing"
)

func TestNewEnv(t *testing.T) {
	os.Setenv("TEST_PRUXY_1", "abc.com=127.0.0.1:8081,127.0.0.8082")
	os.Setenv("TEST_PRUXY_2", "abc.com/abc/123=127.0.0.1,127.0.0.1:8080")
	os.Setenv("TEST_PRUXY_3", "abc.com/abc/123")
	defer os.Unsetenv("TEST_PRUXY_1")
	defer os.Unsetenv("TEST_PRUXY_2")
	defer os.Unsetenv("TEST_PRUXY_3")

	p, err := NewEnv("TEST_PRUXY_")
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
	os.Setenv("TEST_PRUXY_1", "abc.com=127.0.0.1:8081,127.0.0.8082")
	os.Setenv("TEST_PRUXY_2", "abc.com/abc/123=127.0.0.1")
	defer os.Unsetenv("TEST_PRUXY_1")
	defer os.Unsetenv("TEST_PRUXY_2")

	p, err := NewEnv("TEST_PRUXY_")
	if err != nil {
		t.Fatal(err)
	}
	p.DefaultRequestConverter()
}

func TestEnvConvert(t *testing.T) {
	os.Setenv("TEST_PRUXY_1", "abc.com=127.0.0.1:8081,127.0.0.1:8082")
	os.Setenv("TEST_PRUXY_2", "abc.com/abc/123=127.0.0.1")
	defer os.Unsetenv("TEST_PRUXY_1")
	defer os.Unsetenv("TEST_PRUXY_2")

	p, _ := NewEnv("TEST_PRUXY_")
	convertFunc := p.DefaultRequestConverter()

	var convertTests = []struct {
		in  string
		out string
	}{
		{"http://abc.com/a/b/", "http://127.0.0.1:8081/a/b/"},
		{"http://abc.com/a/b", "http://127.0.0.1:8082/a/b"},
		{"http://abc.com/abc/123/456/hi", "http://127.0.0.1/456/hi"},
	}

	for i, tt := range convertTests {
		in, _ := http.NewRequest("GET", tt.in, nil)
		out := copyRequest(in)
		convertFunc(in, out)
		if out.URL.String() != tt.out {
			t.Errorf("%d expected %s, got %s", i, tt.out, out.URL.String())
		}
	}
}
