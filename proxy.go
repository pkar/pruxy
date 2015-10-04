// A modified version from this article.
// http://r7kamura.github.io/2014/07/20/golang-reverse-proxy.html
package pruxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

// These headers won't be copied from original request to proxy request.
var ignoredHeaderNames = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// Proxy provides a host-based proxy server.
type Proxy struct {
	RequestConverter func(originalRequest, pr *http.Request)
	Transport        http.RoundTripper
}

// NewProxyWithHostConverter creates a host-based reverse-proxy.
func NewProxyWithHostConverter(hostConverter func(string) string) *Proxy {
	return &Proxy{
		RequestConverter: func(originalRequest, proxy *http.Request) {
			proxy.URL.Host = hostConverter(originalRequest.Host)
		},
		Transport: http.DefaultTransport,
	}
}

// NewProxyWithRequestConverter creates a request-based reverse-proxy.
func NewProxyWithRequestConverter(requestConverter func(*http.Request, *http.Request)) *Proxy {
	return &Proxy{
		RequestConverter: requestConverter,
		Transport:        http.DefaultTransport,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create a new proxy request object by coping the original request.
	pr := copyRequest(req)

	// Convert an original request into another proxy request.
	p.RequestConverter(req, pr)

	// Convert a request into a response by using its Transport.
	resp, err := p.Transport.RoundTrip(pr)
	if err != nil {
		log.Printf("err: %v %s%s", err, pr.URL.Host, pr.URL.Path)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure a response body from upstream will be always closed.
	defer resp.Body.Close()

	// Copy all header fields.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// copyRequest creates a new proxy request with some modifications from an original request.
func copyRequest(originalRequest *http.Request) *http.Request {
	pr := new(http.Request)
	*pr = *originalRequest
	pr.Proto = "HTTP/1.1"
	pr.ProtoMajor = 1
	pr.ProtoMinor = 1
	pr.Close = false
	pr.Header = make(http.Header)
	pr.URL.Scheme = "http"
	pr.URL.Path = originalRequest.URL.Path

	// Copy all header fields.
	for key, values := range originalRequest.Header {
		for _, value := range values {
			pr.Header.Add(key, value)
		}
	}

	// Remove ignored header fields.
	for _, header := range ignoredHeaderNames {
		pr.Header.Del(header)
	}

	// Append this machine's host name into X-Forwarded-For.
	if requestHost, _, err := net.SplitHostPort(originalRequest.RemoteAddr); err == nil {
		if originalValues, ok := pr.Header["X-Forwarded-For"]; ok {
			requestHost = strings.Join(originalValues, ", ") + ", " + requestHost
		}
		pr.Header.Set("X-Forwarded-For", requestHost)
	}

	return pr
}
