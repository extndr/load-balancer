package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	"github.com/extndr/load-balancer/internal/config"
)

type Proxy struct {
	Client *http.Client
}

func NewProxy(timeout time.Duration, transportConfig config.HTTPTransportConfig) *Proxy {
	return &Proxy{
		Client: &http.Client{
			Transport: NewTransport(transportConfig),
			Timeout:   timeout,
		},
	}
}

func (p *Proxy) DoRequest(b *backend.Backend, r *http.Request) (*http.Response, error) {
	target := &url.URL{
		Scheme:   b.URL.Scheme,
		Host:     b.URL.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header = r.Header.Clone()

	removeHopByHopHeaders(req.Header)
	addForwardedHeaders(req, r)

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to proxy request to %s: %w", target.Host, err)
	}

	removeHopByHopHeaders(resp.Header)

	return resp, nil
}
