package ratelimiter

import (
	"sync"

	"github.com/veyselaksin/strigo/pkg/limiter"
)

// Manager tüm rate limiter'ları yönetir
type Manager struct {
	limiters map[string]limiter.Limiter
	mu       sync.RWMutex
	backend  limiter.Backend
	address  string
}

// NewManager creates a new rate limiter manager
func NewManager(backend limiter.Backend, address string) *Manager {
	return &Manager{
		limiters: make(map[string]limiter.Limiter),
		backend:  backend,
		address:  address,
	}
}

// GetLimiter returns existing limiter or creates new one
func (m *Manager) GetLimiter(cfg limiter.Config) (limiter.Limiter, error) {
	key := cfg.Prefix

	m.mu.RLock()
	if lim, exists := m.limiters[key]; exists {
		m.mu.RUnlock()
		return lim, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if lim, exists := m.limiters[key]; exists {
		return lim, nil
	}

	cfg.Backend = m.backend
	cfg.Address = m.address
	lim, err := limiter.NewLimiter(cfg)
	if err != nil {
		return nil, err
	}

	m.limiters[key] = lim
	return lim, nil
}

// Close closes all limiters
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, lim := range m.limiters {
		if err := lim.Close(); err != nil {
			return err
		}
	}
	return nil
}
