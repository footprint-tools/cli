package dispatchers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandCategory_String(t *testing.T) {
	tests := []struct {
		category CommandCategory
		expected string
	}{
		{CategoryUncategorized, "other commands"},
		{CategoryGetStarted, "get started"},
		{CategoryInspectActivity, "inspect activity and state"},
		{CategoryManageRepos, "manage tracked repositories"},
		{CategoryConfig, "configure fp"},
		{CategoryTheme, "customize appearance"},
		{CategoryPlumbing, "low-level commands (plumbing)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.category.String())
		})
	}
}

func TestCommandCategory_Unknown(t *testing.T) {
	unknownCategory := CommandCategory(99)
	require.Equal(t, "other commands", unknownCategory.String())
}

func TestCategoryOrder(t *testing.T) {
	// Verify category order contains all categories
	require.NotEmpty(t, categoryOrder)
	require.Contains(t, categoryOrder, CategoryGetStarted)
	require.Contains(t, categoryOrder, CategoryInspectActivity)
	require.Contains(t, categoryOrder, CategoryManageRepos)
	require.Contains(t, categoryOrder, CategoryConfig)
	require.Contains(t, categoryOrder, CategoryTheme)
	require.Contains(t, categoryOrder, CategoryPlumbing)
	require.Contains(t, categoryOrder, CategoryUncategorized)
}

func TestCategoryOrderFunction(t *testing.T) {
	// Test the exported CategoryOrder() function
	order := CategoryOrder()
	require.NotEmpty(t, order)
	require.Equal(t, categoryOrder, order)
	require.Contains(t, order, CategoryGetStarted)
	require.Contains(t, order, CategoryInspectActivity)
	require.Contains(t, order, CategoryManageRepos)
}

