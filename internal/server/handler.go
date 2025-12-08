package server

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/extndr/load-balancer/internal/balancer"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	director *balancer.Director
}

func NewHandler(director *balancer.Director) *Handler {
	return &Handler{
		director: director,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := h.director.SelectBackend()
	if target == nil {
		http.Error(w, "Service temporary unavailable", http.StatusServiceUnavailable)
		return
	}

	r.Header.Set("X-Backend", target.URL.Host)

	resp, err := h.director.ProxyRequest(target, r)
	if err != nil {
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Stream response body
	_, err = io.Copy(w, resp.Body)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Errorf("failed to write response body: %v", err)
		return
	}
}
