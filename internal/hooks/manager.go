package hooks

import "github.com/Skryensya/footprint/internal/domain"

// Manager wraps git hooks operations and implements domain.HooksManager.
type Manager struct{}

// NewManager creates a new hooks manager.
func NewManager() *Manager {
	return &Manager{}
}

// Status returns the installation status of hooks at the given path.
func (m *Manager) Status(hooksPath string) (domain.HooksStatus, error) {
	statusMap := Status(hooksPath)

	// Check if any hook is installed
	installed := false
	var managedHooks []string

	for hook, exists := range statusMap {
		if exists {
			installed = true
			managedHooks = append(managedHooks, hook)
		}
	}

	return domain.HooksStatus{
		Installed:    installed,
		HooksPath:    hooksPath,
		ManagedHooks: managedHooks,
	}, nil
}

// Install installs hooks at the given path.
func (m *Manager) Install(hooksPath string) error {
	return Install(hooksPath)
}

// Uninstall removes hooks from the given path.
func (m *Manager) Uninstall(hooksPath string) error {
	return Uninstall(hooksPath)
}

// Verify Manager implements domain.HooksManager
var _ domain.HooksManager = (*Manager)(nil)
