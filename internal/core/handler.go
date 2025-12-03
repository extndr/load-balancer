package core

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Handler struct {
	director *Director
}

func NewHandler(director *Director) *Handler {
	return &Handler{
		director: director,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("incoming request: %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)

	target := h.director.SelectBackend()
	if target == nil {
		log.Errorf("no available backends for request %s %s", r.Method, r.RequestURI)
		http.Error(w, "No alive backends", http.StatusServiceUnavailable)
		return
	}

	start := time.Now()
	resp, err := h.director.ProxyRequest(target, r)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	log.Infof("%s → %s %d %dms",
		r.Method,
		target.URL.Host,
		resp.StatusCode,
		elapsed,
	)

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
