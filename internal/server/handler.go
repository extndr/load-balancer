package server

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/extndr/load-balancer/internal/balancer"
	"github.com/extndr/load-balancer/internal/middleware"
	"github.com/extndr/load-balancer/internal/proxy"
	"go.uber.org/zap"
)

type Handler struct {
	balancer *balancer.Balancer
	proxy    *proxy.Proxy
}

func NewHandler(balancer *balancer.Balancer, proxy *proxy.Proxy) *Handler {
	return &Handler{
		balancer: balancer,
		proxy:    proxy,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := middleware.Logger(r.Context())

	target := h.balancer.NextBackend()
	if target == nil {
		logger.Warn("no healthy backend available")
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	logger = logger.With(zap.String("backend", target.URL.Host))
	ctx := middleware.WithLogger(r.Context(), logger)
	r = r.WithContext(ctx)

	r.Header.Set("X-Backend", target.URL.Host)

	resp, err := h.proxy.DoRequest(target, r)
	if err != nil {
		h.handleProxyError(w, r, err)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Warn("failed to write response body", zap.Error(err))
	}
}

func (h *Handler) handleProxyError(w http.ResponseWriter, r *http.Request, err error) {
	logger := middleware.Logger(r.Context())

	msg := "proxy request failed"
	if errors.Is(err, context.DeadlineExceeded) {
		msg = "proxy request timed out"
	}

	logger.Error(msg, zap.Error(err))
	http.Error(w, "Bad gateway", http.StatusBadGateway)
}
