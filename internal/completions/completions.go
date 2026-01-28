package completions

import "github.com/footprint-tools/cli/internal/dispatchers"

// CommandInfo represents a command extracted from the dispatch tree
type CommandInfo struct {
	Name        string
	Path        []string // Full path from root (e.g., ["fp", "config", "set"])
	Summary     string
	Subcommands []string
	Flags       []FlagInfo
}

// FlagInfo represents a flag for a command
type FlagInfo struct {
	Names       []string
	Description string
	HasValue    bool
}

// ExtractCommands walks the dispatch tree and extracts all commands
func ExtractCommands(root *dispatchers.DispatchNode) []CommandInfo {
	var commands []CommandInfo
	extractNode(root, &commands)
	return commands
}

func extractNode(node *dispatchers.DispatchNode, commands *[]CommandInfo) {
	if node == nil {
		return
	}

	// Extract subcommand names
	var subcommands []string
	for name := range node.Children {
		subcommands = append(subcommands, name)
	}

	// Extract flags
	var flags []FlagInfo
	for _, f := range node.Flags {
		flags = append(flags, FlagInfo{
			Names:       f.Names,
			Description: f.Description,
			HasValue:    f.ValueHint != "",
		})
	}

	cmd := CommandInfo{
		Name:        node.Name,
		Path:        node.Path,
		Summary:     node.Summary,
		Subcommands: subcommands,
		Flags:       flags,
	}
	*commands = append(*commands, cmd)

	// Recurse into children
	for _, child := range node.Children {
		extractNode(child, commands)
	}
}

// FindCommand finds a command by its path
func FindCommand(commands []CommandInfo, path []string) *CommandInfo {
	for i := range commands {
		if pathsEqual(commands[i].Path, path) {
			return &commands[i]
		}
	}
	return nil
}

func pathsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
