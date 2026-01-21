package dispatchers

import (
	"sort"
	"strings"
)

// levenshtein calculates the edit distance between two strings
func levenshtein(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize first column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}

	// Initialize first row
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

type suggestion struct {
	name     string
	distance int
}

// FindSimilarCommands finds commands similar to the input string
// It searches in the given node's children and returns up to maxResults suggestions
func FindSimilarCommands(input string, node *DispatchNode, maxResults int) []string {
	if node == nil || node.Children == nil {
		return nil
	}

	const maxDistance = 3

	var suggestions []suggestion

	for name := range node.Children {
		dist := levenshtein(input, name)
		if dist <= maxDistance && dist > 0 {
			suggestions = append(suggestions, suggestion{name: name, distance: dist})
		}
	}

	// Sort by distance (ascending), then alphabetically for stability
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].distance != suggestions[j].distance {
			return suggestions[i].distance < suggestions[j].distance
		}
		return suggestions[i].name < suggestions[j].name
	})

	// Limit results
	if len(suggestions) > maxResults {
		suggestions = suggestions[:maxResults]
	}

	// Extract names
	result := make([]string, len(suggestions))
	for i, s := range suggestions {
		result[i] = s.name
	}

	return result
}

// CollectAllCommands recursively collects all command names from a node tree
// This can be used for global command suggestions
func CollectAllCommands(node *DispatchNode, prefix string) []string {
	if node == nil {
		return nil
	}

	var commands []string

	for name, child := range node.Children {
		fullPath := name
		if prefix != "" {
			fullPath = prefix + " " + name
		}
		commands = append(commands, fullPath)
		commands = append(commands, CollectAllCommands(child, fullPath)...)
	}

	return commands
}
