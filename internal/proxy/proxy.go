package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/extndr/load-balancer/internal/backend"
	"github.com/extndr/load-balancer/internal/config"
	log "github.com/sirupsen/logrus"
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
		log.Errorf("failed to create request to %s: %v", b.URL.Host, err)
		return nil, err
	}

	req.Header = r.Header.Clone()

	removeHopByHopHeaders(req.Header)
	addForwardedHeaders(req, r)

	resp, err := p.Client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warnf("proxy timeout: backend %s did not respond within %v", b.URL.Host, p.Client.Timeout)
		} else {
			log.Errorf("proxy request to %s failed: %v", b.URL.Host, err)
		}
		return nil, err
	}

	removeHopByHopHeaders(resp.Header)

	return resp, nil
}
