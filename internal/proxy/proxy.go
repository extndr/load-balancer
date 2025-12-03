package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/extndr/load-balancer/internal/config"
	"github.com/extndr/load-balancer/internal/pool"
	log "github.com/sirupsen/logrus"

	httputil "github.com/extndr/load-balancer/internal/http"
)

type Proxy struct {
	Client  *http.Client
	Timeout time.Duration
}

func NewProxy(timeout time.Duration, transportConfig config.HTTPTransportConfig) *Proxy {
	return &Proxy{
		Client: &http.Client{
			Transport: httputil.NewTransport(transportConfig),
			Timeout:   timeout,
		},
	}
}

func (p *Proxy) DoRequest(b *pool.Backend, r *http.Request) (*http.Response, error) {
	target := &url.URL{
		Scheme:   b.URL.Scheme,
		Host:     b.URL.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	ctx, cancel := context.WithTimeout(r.Context(), p.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, target.String(), r.Body)
	if err != nil {
		log.Errorf("failed to create request to %s: %v", b.URL.Host, err)
		return nil, err
	}

	req.Header = r.Header.Clone()

	resp, err := p.Client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warnf("proxy timeout: backend %s did not respond within %v", b.URL.Host, p.Timeout)
		} else {
			log.Errorf("proxy request to %s failed: %v", b.URL.Host, err)
		}
		return nil, err
	}

	return resp, nil
}
