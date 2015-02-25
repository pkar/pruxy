package pruxy

import (
	"strings"
	"testing"
)

func TestNewEtcd(t *testing.T) {
	etcdHostList := strings.Split("127.0.0.1:4001,127.0.0.1:4002", ",")
	_, err := NewEtcd(etcdHostList, "pruxy")
	if err != nil {
		t.Fatal(err)
	}
}
