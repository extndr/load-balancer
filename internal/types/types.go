package types

import "time"

type StatusEvent struct {
	URL     string
	Healthy bool
	Latency time.Duration
	Error   error
}
