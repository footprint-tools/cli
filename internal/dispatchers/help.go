package dispatchers

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/footprint-tools/cli/internal/help"
	"github.com/footprint-tools/cli/internal/ui"
	"github.com/footprint-tools/cli/internal/ui/style"
)

// commandDisplayOrder defines explicit ordering within categories.
// Commands not listed appear alphabetically after listed ones.
var commandDisplayOrder = map[string]int{
	// get started
	"setup": 1,
	// inspect activity and state
	"activity": 1,
	"watch":    2,
	"repos":    3,
	"list":     4,
	"check":    5,
	"version":  6,
	"logs":     7,
	// manage tracked repositories
	"teardown": 1,
	// config commands
	"config get":   1,
	"config set":   2,
	"config unset": 3,
	"config list":  4,
	// theme commands
	"theme list": 1,
	"theme set":  2,
	"theme pick": 3,
}

// formatUsage styles the usage line with the command in Info color and the rest muted.
func formatUsage(usage string) string {
	// Find where the command ends (first [ or <)
	cmdEnd := len(usage)
	for i, c := range usage {
		if c == '[' || c == '<' {
			cmdEnd = i
			break
		}
	}

	cmd := strings.TrimSpace(usage[:cmdEnd])
	rest := ""
	if cmdEnd < len(usage) {
		rest = usage[cmdEnd:]
	}

	if rest == "" {
		return style.Info(cmd)
	}
	return style.Info(cmd) + " " + style.Muted(rest)
}

func collectLeafCommands(node *DispatchNode, out *[]*DispatchNode) {
	if node.Action != nil {
		*out = append(*out, node)
		return
	}

	for _, child := range node.Children {
		collectLeafCommands(child, out)
	}
}

// HelpAction generates help output for a command node.
func HelpAction(node *DispatchNode, root *DispatchNode) CommandFunc {
	return func(args []string, flags *ParsedFlags) error {
		var out bytes.Buffer

		if node == root {
			// Root help: git-like format
			out.WriteString("fp - ")
			out.WriteString(node.Summary)
			out.WriteString("\n\n")

			out.WriteString("USAGE\n   ")
			out.WriteString(formatUsage(node.Usage))
			out.WriteString("\n\n")

			grouped := make(map[CommandCategory][]*DispatchNode)

			var leaves []*DispatchNode
			for _, child := range root.Children {
				collectLeafCommands(child, &leaves)
			}

			for _, cmd := range leaves {
				grouped[cmd.Category] = append(grouped[cmd.Category], cmd)
			}

			for _, cat := range categoryOrder {
				cmds := grouped[cat]
				if len(cmds) == 0 {
					continue
				}

				out.WriteString(cat.String())
				out.WriteString("\n")

				// Sort by explicit order, then alphabetically
				sort.Slice(cmds, func(i, j int) bool {
					nameI := strings.Join(cmds[i].Path[1:], " ")
					nameJ := strings.Join(cmds[j].Path[1:], " ")
					orderI, hasI := commandDisplayOrder[nameI]
					orderJ, hasJ := commandDisplayOrder[nameJ]
					if hasI && hasJ {
						return orderI < orderJ
					}
					if hasI {
						return true
					}
					if hasJ {
						return false
					}
					return nameI < nameJ
				})

				for _, cmd := range cmds {
					displayName := strings.Join(cmd.Path[1:], " ")
					fmt.Fprintf(&out, "   %s  %s\n", style.Info(fmt.Sprintf("%-16s", displayName)), cmd.Summary)
				}
				out.WriteString("\n")
			}

			// Conceptual guides section
			out.WriteString("conceptual guides\n")
			for _, topic := range help.AllTopics() {
				fmt.Fprintf(&out, "   %s  %s\n", style.Muted(fmt.Sprintf("%-16s", topic.Name)), topic.Summary)
			}
			out.WriteString("\n")

			// Footer
			out.WriteString("See 'fp help <command>' for detailed help on a specific command.\n")
			out.WriteString("See 'fp help <topic>' for conceptual documentation.\n")
		} else {
			// Subcommand help
			out.WriteString(strings.Join(node.Path, " "))
			if node.Summary != "" {
				out.WriteString(" - ")
				out.WriteString(node.Summary)
			}
			out.WriteString("\n\n")

			out.WriteString("USAGE\n   ")
			out.WriteString(formatUsage(node.Usage))
			out.WriteString("\n\n")

			// Show longer description if available
			if node.Description != "" {
				out.WriteString(node.Description)
				out.WriteString("\n\n")
			}

			if len(node.Children) > 0 {
				out.WriteString("COMMANDS\n")

				children := make([]*DispatchNode, 0, len(node.Children))
				for _, child := range node.Children {
					children = append(children, child)
				}

				sort.Slice(children, func(i, j int) bool {
					pathI := strings.Join(children[i].Path[1:], " ")
					pathJ := strings.Join(children[j].Path[1:], " ")
					orderI, hasI := commandDisplayOrder[pathI]
					orderJ, hasJ := commandDisplayOrder[pathJ]
					if hasI && hasJ {
						return orderI < orderJ
					}
					if hasI {
						return true
					}
					if hasJ {
						return false
					}
					return children[i].Name < children[j].Name
				})

				for _, child := range children {
					fmt.Fprintf(&out, "   %s  %s\n", style.Info(fmt.Sprintf("%-12s", child.Name)), child.Summary)
				}
				out.WriteString("\n")
			}

			if len(node.Flags) > 0 {
				out.WriteString("FLAGS\n")
				for _, f := range node.Flags {
					name := strings.Join(f.Names, ", ")
					if f.ValueHint != "" {
						name = name + " " + f.ValueHint
					}
					fmt.Fprintf(&out, "   %s  %s\n", style.Info(fmt.Sprintf("%-24s", name)), f.Description)
				}
				out.WriteString("\n")
			}

			out.WriteString("See 'fp help <command>' to read about a specific command.\n")
		}

		ui.Pager(out.String())
		return nil
	}
}

// TopicHelpAction generates help output for a conceptual topic.
func TopicHelpAction(topic *help.Topic) CommandFunc {
	return func(args []string, flags *ParsedFlags) error {
		ui.Pager(topic.Content())
		return nil
	}
}

// TopicsListAction lists all available help topics.
func TopicsListAction() CommandFunc {
	return func(args []string, flags *ParsedFlags) error {
		var out bytes.Buffer

		out.WriteString("TOPICS\n\n")

		for _, topic := range help.AllTopics() {
			fmt.Fprintf(&out, "   %s  %s\n", style.Muted(fmt.Sprintf("%-12s", topic.Name)), topic.Summary)
		}

		out.WriteString("\nSee 'fp help <topic>' to read about a specific topic.\n")

		ui.Pager(out.String())
		return nil
	}
}
