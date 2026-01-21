package config

import "github.com/Skryensya/footprint/internal/domain"

// Provider wraps configuration operations and implements domain.ConfigProvider.
type Provider struct{}

// NewProvider creates a new configuration provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Get returns the value for a configuration key.
func (p *Provider) Get(key string) (string, bool) {
	return Get(key)
}

// GetAll returns all configuration values.
func (p *Provider) GetAll() (map[string]string, error) {
	return GetAll()
}

// Set sets a configuration value.
func (p *Provider) Set(key, value string) error {
	lines, err := ReadLines()
	if err != nil {
		return err
	}

	lines, _ = Set(lines, key, value)
	return WriteLines(lines)
}

// Unset removes a configuration value.
func (p *Provider) Unset(key string) error {
	lines, err := ReadLines()
	if err != nil {
		return err
	}

	lines, _ = Unset(lines, key)
	return WriteLines(lines)
}

// Verify Provider implements domain.ConfigProvider
var _ domain.ConfigProvider = (*Provider)(nil)
