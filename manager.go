package strigo

import (
	"sync"
)

// Manager manages all rate limiters
type Manager struct {
	limiters map[string]Limiter
	mu       sync.RWMutex
	backend  Backend
	address  string
}

// NewManager creates a new rate limiter manager
func NewManager(backend Backend, address string) *Manager {
	return &Manager{
		limiters: make(map[string]Limiter),
		backend:  backend,
		address:  address,
	}
}

// GetLimiter returns existing limiter or creates new one
func (m *Manager) GetLimiter(cfg LimiterConfig) (Limiter, error) {
	// Create a unique key based on config
	key := cfg.GetUniqueKey()

	m.mu.RLock()
	if lim, exists := m.limiters[key]; exists {
		m.mu.RUnlock()
		return lim, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if lim, exists := m.limiters[key]; exists {
		return lim, nil
	}

	// Set backend configuration
	cfg.Backend = m.backend
	cfg.Address = m.address

	// Create new limiter
	lim, err := NewLimiter(cfg)
	if err != nil {
		return nil, err
	}

	m.limiters[key] = lim
	return lim, nil
}

// Allow checks if a request should be allowed
func (m *Manager) Allow(key string, cfg LimiterConfig) bool {
	lim, err := m.GetLimiter(cfg)
	if err != nil {
		return false
	}
	return lim.Allow(key)
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
