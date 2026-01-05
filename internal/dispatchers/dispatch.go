package dispatchers

import (
	"errors"
	"footprint/internal/usage"
	"strings"
)

func Dispatch(root *DispatchNode, tokens []string, flags []string) (Resolution, error) {
	current := root
	lastValid := root
	pathLen := 0

	// ──────────────────────────────
	// 1. Detect "help" command first
	// ──────────────────────────────
	for i, tok := range tokens {
		if tok == "help" {
			targetPath := tokens[:i]
			if len(tokens[i+1:]) > 0 {
				targetPath = tokens[i+1:]
			}

			target := resolveNode(root, targetPath)
			if target == nil {
				return Resolution{}, usage.UnknownCommand(strings.Join(targetPath, " "))
			}

			return Resolution{
				Node:    target,
				Args:    nil,
				Flags:   flags,
				Execute: HelpAction(target, root),
			}, nil
		}
	}

	// ──────────────────────────────
	// 2. Traverse command path
	// ──────────────────────────────
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

	// ──────────────────────────────
	// 3. --help / -h always wins
	// ──────────────────────────────
	if hasHelpFlag(flags) {
		return Resolution{
			Node:    lastValid,
			Args:    nil,
			Flags:   flags,
			Execute: HelpAction(lastValid, root),
		}, nil
	}

	// ──────────────────────────────
	// 4. Validate flags
	// ──────────────────────────────
	valid := validFlagsForNode(current, root)
	if err := validateFlags(flags, valid); err != nil {
		return Resolution{}, err
	}

	// ──────────────────────────────
	// 5. Validate arguments
	// ──────────────────────────────
	if err := validateArgs(current.Args, args); err != nil {
		return Resolution{}, err
	}

	// ──────────────────────────────
	// 6. Non-executable node → help
	// ──────────────────────────────
	if current.Action == nil {
		return Resolution{
			Node:    current,
			Args:    nil,
			Flags:   flags,
			Execute: HelpAction(current, root),
		}, nil
	}

	// ──────────────────────────────
	// 7. Execute
	// ──────────────────────────────
	return Resolution{
		Node:    current,
		Args:    args,
		Flags:   flags,
		Execute: current.Action,
	}, nil
}

func hasHelpFlag(flags []string) bool {
	for _, f := range flags {
		if f == "--help" || f == "-h" {
			return true
		}
	}
	return false
}

func validFlagsForNode(node *DispatchNode, root *DispatchNode) map[string]bool {
	valid := make(map[string]bool)

	// global flags (root)
	for _, f := range root.Flags {
		for _, name := range f.Names {
			valid[name] = true
		}
	}

	// local flags
	for _, f := range node.Flags {
		for _, name := range f.Names {
			valid[name] = true
		}
	}

	return valid
}

func validateFlags(flags []string, valid map[string]bool) error {
	for _, f := range flags {
		if !valid[f] {
			return errors.New("#TODO: not valid flags")
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
