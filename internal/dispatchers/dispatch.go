package dispatchers

import (
	"strings"
	"sync"

	"github.com/footprint-tools/cli/internal/help"
	"github.com/footprint-tools/cli/internal/usage"
)

const defaultSuggestionsCount = 3

// interactiveBrowserFunc is injected from main to avoid import cycles
var (
	interactiveBrowserFunc   CommandFunc
	interactiveBrowserFuncMu sync.RWMutex
)

// SetInteractiveBrowserFunc sets the interactive browser function thread-safely.
func SetInteractiveBrowserFunc(fn CommandFunc) {
	interactiveBrowserFuncMu.Lock()
	defer interactiveBrowserFuncMu.Unlock()
	interactiveBrowserFunc = fn
}

// getInteractiveBrowserFunc gets the interactive browser function thread-safely.
func getInteractiveBrowserFunc() CommandFunc {
	interactiveBrowserFuncMu.RLock()
	defer interactiveBrowserFuncMu.RUnlock()
	return interactiveBrowserFunc
}

func handleHelpCommand(root *DispatchNode, tokens []string, flags *ParsedFlags) (Resolution, error, bool) {
	for i, tok := range tokens {
		if tok != "help" {
			continue
		}

		browserFn := getInteractiveBrowserFunc()
		if hasInteractiveFlag(flags) && browserFn != nil {
			return Resolution{Node: root, Flags: flags, Execute: browserFn}, nil, true
		}

		targetPath := tokens[:i]
		if len(tokens[i+1:]) > 0 {
			targetPath = tokens[i+1:]
		}

		if len(targetPath) == 1 && targetPath[0] == "topics" {
			return Resolution{Node: root, Flags: flags, Execute: TopicsListAction()}, nil, true
		}

		target := resolveNode(root, targetPath)
		if target != nil {
			return Resolution{Node: target, Flags: flags, Execute: HelpAction(target, root)}, nil, true
		}

		if len(targetPath) == 1 {
			topic := help.LookupTopic(targetPath[0])
			if topic != nil {
				return Resolution{Node: root, Flags: flags, Execute: TopicHelpAction(topic)}, nil, true
			}
		}

		suggestions := FindSimilarCommands(targetPath[0], root, defaultSuggestionsCount)
		return Resolution{}, usage.UnknownCommand(strings.Join(targetPath, " "), suggestions...), true
	}
	return Resolution{}, nil, false
}

func Dispatch(root *DispatchNode, tokens []string, flags *ParsedFlags) (Resolution, error) {
	if res, err, handled := handleHelpCommand(root, tokens, flags); handled {
		return res, err
	}

	current := root
	lastValid := root
	pathLen := 0

	for i, tok := range tokens {
		child, ok := current.Children[tok]
		if !ok {
			// If this is the first token (top-level command) and it doesn't exist,
			// it's likely a typo - suggest similar commands
			if i == 0 && len(current.Children) > 0 {
				suggestions := FindSimilarCommands(tok, current, defaultSuggestionsCount)
				return Resolution{}, usage.UnknownCommand(tok, suggestions...)
			}
			// For subcommands: if the current node has children but no action,
			// it's a command group and unknown tokens are errors, not args
			if current.Action == nil && len(current.Children) > 0 {
				suggestions := FindSimilarCommands(tok, current, defaultSuggestionsCount)
				cmdPath := strings.Join(append(current.Path, tok), " ")
				return Resolution{}, usage.UnknownCommand(cmdPath, suggestions...)
			}
			break
		}
		current = child
		lastValid = child
		pathLen++
	}

	args := tokens[pathLen:]

	if hasHelpFlag(flags) {
		return Resolution{
			Node:    lastValid,
			Args:    nil,
			Flags:   flags,
			Execute: HelpAction(lastValid, root),
		}, nil
	}

	valid := validFlagsForNode(current, root)
	if err := validateFlags(flags, valid); err != nil {
		return Resolution{}, err
	}

	if err := validateArgs(current.Args, args); err != nil {
		return Resolution{}, err
	}

	if current.Action == nil {
		// Check for interactive flag first
		if hasInteractiveFlag(flags) && current.InteractiveAction != nil {
			return Resolution{
				Node:    current,
				Args:    args,
				Flags:   flags,
				Execute: current.InteractiveAction,
			}, nil
		}
		// No command specified: show help but exit with code 1 (like git)
		exitCode := 0
		if current == root && len(tokens) == 0 {
			exitCode = 1
		}
		return Resolution{
			Node:     current,
			Args:     nil,
			Flags:    flags,
			Execute:  HelpAction(current, root),
			ExitCode: exitCode,
		}, nil
	}

	return Resolution{
		Node:    current,
		Args:    args,
		Flags:   flags,
		Execute: current.Action,
	}, nil
}

func hasHelpFlag(flags *ParsedFlags) bool {
	return flags.Has("--help") || flags.Has("-h")
}

func hasInteractiveFlag(flags *ParsedFlags) bool {
	return flags.Has("--interactive") || flags.Has("-i")
}

func validFlagsForNode(node *DispatchNode, root *DispatchNode) map[string]bool {
	valid := make(map[string]bool)

	for _, f := range root.Flags {
		for _, name := range f.Names {
			valid[name] = true
		}
	}

	for _, f := range node.Flags {
		for _, name := range f.Names {
			valid[name] = true
		}
	}

	return valid
}

func validateFlags(flags *ParsedFlags, valid map[string]bool) error {
	for _, f := range flags.Raw() {
		// Extract the flag name (strip value after =)
		name := f
		if idx := strings.Index(f, "="); idx != -1 {
			name = f[:idx]
		}
		if !valid[name] {
			return usage.InvalidFlag(f)
		}
	}
	return nil
}

func validateArgs(spec []ArgSpec, args []string) error {
	requiredCount := 0
	for _, a := range spec {
		if a.Required {
			requiredCount++
		}
	}

	if len(args) < requiredCount {
		// Bounds check: ensure spec has enough elements
		if len(args) >= len(spec) {
			return usage.MissingArgument("argument")
		}
		missing := spec[len(args)].Name
		return usage.MissingArgument(missing)
	}

	return nil
}

func resolveNode(root *DispatchNode, path []string) *DispatchNode {
	current := root

	for _, p := range path {
		child, ok := current.Children[p]
		if !ok {
			return nil
		}
		current = child
	}

	return current
}
