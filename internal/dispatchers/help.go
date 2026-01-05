package dispatchers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func page(content string) error {
	cmd := exec.Command("less", "-FRSX")
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func HelpAction(node *DispatchNode, root *DispatchNode) CommandFunc {
	return func(args []string, flags []string) error {
		var out bytes.Buffer

		// ──────────────────────────────
		// HEADER
		// ──────────────────────────────
		out.WriteString(strings.Join(node.Path, " "))
		if node.Summary != "" {
			out.WriteString(" - ")
			out.WriteString(node.Summary)
		}
		out.WriteString("\n\n")

		// ──────────────────────────────
		// USAGE
		// ──────────────────────────────
		out.WriteString("usage: ")
		out.WriteString(node.Usage)
		out.WriteString("\n\n")

		// ──────────────────────────────
		// ROOT HELP (git-style overview)
		// ──────────────────────────────
		if node == root {
			out.WriteString("These are common fp commands used in various situations:\n\n")

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

				sort.Slice(cmds, func(i, j int) bool {
					return strings.Join(cmds[i].Path[1:], " ") <
						strings.Join(cmds[j].Path[1:], " ")
				})

				for _, cmd := range cmds {
					displayName := strings.Join(cmd.Path[1:], " ")

					out.WriteString(fmt.Sprintf(
						"   %-18s %s\n",
						displayName,
						cmd.Summary,
					))
				}
				out.WriteString("\n")
			}
		}

		// ──────────────────────────────
		// NON-ROOT: show child commands
		// ──────────────────────────────
		if node != root && len(node.Children) > 0 {
			out.WriteString("COMMANDS\n")

			children := make([]*DispatchNode, 0, len(node.Children))
			for _, child := range node.Children {
				children = append(children, child)
			}

			sort.Slice(children, func(i, j int) bool {
				return children[i].Name < children[j].Name
			})

			for _, child := range children {
				out.WriteString(fmt.Sprintf(
					"   %-12s %s\n",
					child.Name,
					child.Summary,
				))
			}
			out.WriteString("\n")
		}

		// ──────────────────────────────
		// FLAGS (local)
		// ──────────────────────────────
		if len(node.Flags) > 0 {
			out.WriteString("FLAGS\n")
			for _, f := range node.Flags {
				out.WriteString(fmt.Sprintf(
					"   %-12s %s\n",
					strings.Join(f.Names, ", "),
					f.Description,
				))
			}
			out.WriteString("\n")
		}

		// ──────────────────────────────
		// FOOTER
		// ──────────────────────────────
		out.WriteString(
			"See 'fp help <command>' to read about a specific command.\n",
		)

		return page(out.String())
	}
}
