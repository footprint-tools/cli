package help

import (
	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/help"
)

type Deps struct {
	BuildTree func() *dispatchers.DispatchNode
	AllTopics func() []*help.Topic
}

// buildTreeFunc is set at runtime to avoid import cycles.
// It is set by SetBuildTreeFunc before Browser is called.
var buildTreeFunc func() *dispatchers.DispatchNode

// SetBuildTreeFunc allows the CLI package to inject the BuildTree function.
func SetBuildTreeFunc(f func() *dispatchers.DispatchNode) {
	buildTreeFunc = f
}

func DefaultDeps() Deps {
	return Deps{
		BuildTree: buildTreeFunc,
		AllTopics: help.AllTopics,
	}
}
