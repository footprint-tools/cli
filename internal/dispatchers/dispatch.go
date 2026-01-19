package dispatchers

import (
	"strings"

	"github.com/Skryensya/footprint/internal/help"
	"github.com/Skryensya/footprint/internal/usage"
)

func Dispatch(root *DispatchNode, tokens []string, flags *ParsedFlags) (Resolution, error) {
	current := root
	lastValid := root
	pathLen := 0

	for i, tok := range tokens {
		if tok == "help" {
			targetPath := tokens[:i]
			if len(tokens[i+1:]) > 0 {
				targetPath = tokens[i+1:]
			}

			// Handle "fp help topics" - list all topics
			if len(targetPath) == 1 && targetPath[0] == "topics" {
				return Resolution{
					Node:    root,
					Args:    nil,
					Flags:   flags,
					Execute: TopicsListAction(),
				}, nil
			}

			// Try to resolve as a command first
			target := resolveNode(root, targetPath)
			if target != nil {
				return Resolution{
					Node:    target,
					Args:    nil,
					Flags:   flags,
					Execute: HelpAction(target, root),
				}, nil
			}

			// Try to resolve as a help topic
			if len(targetPath) == 1 {
				topic := help.LookupTopic(targetPath[0])
				if topic != nil {
					return Resolution{
						Node:    root,
						Args:    nil,
						Flags:   flags,
						Execute: TopicHelpAction(topic),
					}, nil
				}
			}

			// Neither command nor topic found
			return Resolution{}, usage.UnknownCommand(strings.Join(targetPath, " "))
		}
	}

	for _, tok := range tokens {
		child, ok := current.Children[tok]
		if !ok {
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
