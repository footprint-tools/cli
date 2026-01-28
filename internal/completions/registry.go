package completions

import (
	"os"
	"path/filepath"

	"github.com/footprint-tools/cli/internal/dispatchers"
)

var commandTree *dispatchers.DispatchNode
var binaryPath string
var binaryName string

// RegisterCommandTree stores the command tree for later use by completion generators
// This should be called from main.go after building the tree
func RegisterCommandTree(root *dispatchers.DispatchNode) {
	commandTree = root

	// Get the actual executable path
	if exe, err := os.Executable(); err == nil {
		// Resolve symlinks to get the real path
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			binaryPath = resolved
		} else {
			binaryPath = exe
		}
		binaryName = filepath.Base(binaryPath)
	} else if len(os.Args) > 0 {
		binaryPath = os.Args[0]
		binaryName = filepath.Base(os.Args[0])
	}

	if binaryName == "" {
		binaryName = "fp"
		binaryPath = "fp"
	}
}

// GetCommandTree returns the registered command tree
func GetCommandTree() *dispatchers.DispatchNode {
	return commandTree
}

// GetBinaryName returns the name of the binary (e.g., "fp" or "fpdev")
func GetBinaryName() string {
	if binaryName == "" {
		return "fp"
	}
	return binaryName
}

// GetBinaryPath returns the full path to the binary
func GetBinaryPath() string {
	if binaryPath == "" {
		return "fp"
	}
	return binaryPath
}
