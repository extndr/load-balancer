package proxy

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

var hopByHop = map[string]struct{}{
	"Connection":          {},
	"Proxy-Connection":    {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"TE":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func removeHopByHopHeaders(h http.Header) {
	if conn := h.Get("Connection"); conn != "" {
		for token := range strings.SplitSeq(conn, ",") {
			if token = strings.TrimSpace(token); token != "" {
				h.Del(token)
			}
		}
	}

	for name := range hopByHop {
		h.Del(name)
	}
}

func addForwardedHeaders(req, original *http.Request) {
	clientIP, _, _ := net.SplitHostPort(original.RemoteAddr)
	if clientIP == "" {
		clientIP = original.RemoteAddr
	}

	scheme := "http"
	if original.TLS != nil {
		scheme = "https"
	}

	forwarded := fmt.Sprintf(
		"for=%s;proto=%s;host=%s;by=load-balancer",
		clientIP, scheme, original.Host,
	)
	req.Header.Set("Forwarded", forwarded)

	req.Header.Set("X-Forwarded-For", clientIP)
	req.Header.Set("X-Forwarded-Proto", scheme)
	req.Header.Set("X-Forwarded-Host", original.Host)
}
