package pruxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	EtcdServer *httptest.Server
)

// EtcdResponse http://developer.android.com/guide/google/gcm/gcm.html#send-msg
type EtcdResponse struct {
	StatusCode int    `json:"StatusCode"`
	Body       []byte `json:"Body"`
}

func init() {
	EtcdServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := EtcdResponse{
		//StatusCode: 200,
		}
		j, err := json.Marshal(resp)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(w, string(j))
	}))
}

func TestNewEtcd(t *testing.T) {
	etcdHostList := []string{EtcdServer.URL}
	_, err := NewEtcd(etcdHostList, "pruxy")
	if err != nil {
		//t.Fatal(err)
	}
}

func TestEtcdDefaultRequestConverter(t *testing.T) {
	etcdHostList := []string{EtcdServer.URL}
	p, err := NewEtcd(etcdHostList, "pruxy")
	if err != nil {
		//t.Fatal(err)
	}
	p.DefaultRequestConverter()
}

func TestEtcdConvert(t *testing.T) {
	etcdHostList := []string{EtcdServer.URL}
	_, err := NewEtcd(etcdHostList, "pruxy")
	if err != nil {
		//t.Fatal(err)
	}
	//convertFunc := p.DefaultRequestConverter()
	//in, _ := http.NewRequest("GET", "http://"+EtcdServer.URL, nil)
	//out := copyRequest(in)
	//convertFunc(in, out)
}
