package dispatchers

type CommandCategory int

const (
	CategoryUncategorized CommandCategory = iota
	CategoryGetStarted        // First steps: setup, track
	CategoryInspectActivity   // Viewing activity and state
	CategoryManageRepos       // Managing tracked repositories
	CategoryConfig            // Configuration
	CategoryTheme             // Theme customization
	CategoryPlumbing          // Low-level/plumbing commands (record)
)

func (c CommandCategory) String() string {
	switch c {
	case CategoryGetStarted:
		return "get started"
	case CategoryInspectActivity:
		return "inspect activity and state"
	case CategoryManageRepos:
		return "manage tracked repositories"
	case CategoryConfig:
		return "configure fp"
	case CategoryTheme:
		return "customize appearance"
	case CategoryPlumbing:
		return "low-level commands (plumbing)"
	default:
		return "other commands"
	}
}

var categoryOrder = []CommandCategory{
	CategoryGetStarted,
	CategoryInspectActivity,
	CategoryManageRepos,
	CategoryConfig,
	CategoryTheme,
	CategoryPlumbing,
	CategoryUncategorized,
}

// CategoryOrder returns the display order for categories.
func CategoryOrder() []CommandCategory {
	return categoryOrder
}
