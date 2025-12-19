package pool

import (
	"testing"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name         string
		urls         []string
		wantErr      bool
		wantBackends int
	}{
		{"valid URLs", []string{"http://localhost:8080", "https://example.com"}, false, 2},
		{"no URLs", []string{}, true, 0},
		{"invalid URL", []string{"http://invalid url with spaces"}, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPool(tt.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("error: got %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && len(p.GetAll()) != tt.wantBackends {
				t.Errorf("backends: got %d, want %d", len(p.GetAll()), tt.wantBackends)
			}
		})
	}
}

func TestPoolGetAll(t *testing.T) {
	urls := []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
	p, err := NewPool(urls)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	all := p.GetAll()
	if len(all) != len(urls) {
		t.Errorf("GetAll(): got %d backends, want %d", len(all), len(urls))
	}

	for i, backend := range all {
		if backend.URL.String() != urls[i] {
			t.Errorf("backend %d: got %s, want %s", i, backend.URL.String(), urls[i])
		}
	}
}

func TestPoolGetHealthy(t *testing.T) {
	tests := []struct {
		name        string
		setHealthy  []bool
		wantHealthy int
	}{
		{"all healthy", []bool{true, true}, 2},
		{"one unhealthy", []bool{false, true}, 1},
		{"all unhealthy", []bool{false, false}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPool([]string{"http://localhost:8080", "http://localhost:8081"})
			if err != nil {
				t.Fatalf("failed to create pool: %v", err)
			}

			all := p.GetAll()
			for i, healthy := range tt.setHealthy {
				p.UpdateHealth(all[i], healthy)
			}

			got := p.GetHealthy()
			if len(got) != tt.wantHealthy {
				t.Errorf("healthy backends: got %d, want %d", len(got), tt.wantHealthy)
			}
		})
	}
}

func TestPoolUpdateHealth(t *testing.T) {
	tests := []struct {
		name       string
		markHealth bool
		wantCount  int
	}{
		{"mark unhealthy", false, 1},
		{"mark healthy", true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPool([]string{"http://localhost:8080", "http://localhost:8081"})
			if err != nil {
				t.Fatalf("failed to create pool: %v", err)
			}

			backend := p.GetAll()[0]
			p.UpdateHealth(backend, tt.markHealth)

			healthy := p.GetHealthy()
			if len(healthy) != tt.wantCount {
				t.Errorf("healthy backends: got %d, want %d", len(healthy), tt.wantCount)
			}
		})
	}
}
