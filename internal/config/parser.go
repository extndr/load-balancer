package config

import (
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func parseBackends(raw string) []string {
	var backends []string
	for b := range strings.SplitSeq(raw, ",") {
		if trimmed := strings.TrimSpace(b); trimmed != "" {
			backends = append(backends, trimmed)
		}
	}
	return backends
}

func parseDuration(raw string) time.Duration {
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Warnf("invalid duration %q, using default 10s", raw)
		return 10 * time.Second
	}
	return d
}

func parseInt(raw string) int {
	i, err := strconv.Atoi(raw)
	if err != nil {
		log.Warnf("invalid int %q, using default 30", raw)
		return 30
	}
	return i
}
