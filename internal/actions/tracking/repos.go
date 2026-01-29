package tracking

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/hooks"
	"github.com/footprint-tools/cli/internal/store"
	"github.com/footprint-tools/cli/internal/ui/style"
)

// ReposList lists repositories with recorded activity.
func ReposList(_ []string, flags *dispatchers.ParsedFlags) error {
	jsonOutput := flags.Has("--json")
	return reposListImpl(jsonOutput, reposDeps{
		DBPath:    store.DBPath,
		OpenStore: store.New,
		Println:   defaultPrintln,
	})
}

type reposDeps struct {
	DBPath    func() string
	OpenStore func(string) (*store.Store, error)
	Println   func(...any) (int, error)
}

func reposListImpl(jsonOutput bool, deps reposDeps) error {
	s, err := deps.OpenStore(deps.DBPath())
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	repos, err := s.ListRepos()
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		if jsonOutput {
			_, _ = deps.Println("[]")
		} else {
			_, _ = deps.Println("no tracked repositories")
			_, _ = deps.Println("run 'fp setup' in a repo to install hooks")
		}
		return nil
	}

	if jsonOutput {
		type repoJSON struct {
			Path     string `json:"path"`
			AddedAt  string `json:"added_at,omitempty"`
			LastSeen string `json:"last_seen,omitempty"`
		}
		out := make([]repoJSON, 0, len(repos))
		for _, r := range repos {
			out = append(out, repoJSON{Path: r.Path, AddedAt: r.AddedAt, LastSeen: r.LastSeen})
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		_, _ = deps.Println(string(data))
		return nil
	}

	for _, r := range repos {
		_, _ = deps.Println(r.Path)
	}

	return nil
}

// ReposScan scans directories for git repositories and shows their hook status.
func ReposScan(_ []string, flags *dispatchers.ParsedFlags) error {
	jsonOutput := flags.Has("--json")
	root := flags.String("--root", ".")

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", root, err)
	}
	root = absRoot

	maxDepth := flags.Int("--depth", 25)

	if !jsonOutput {
		fmt.Printf("Scanning for git repositories in %s...\n", root)
	}
	repos, err := scanForRepos(root, maxDepth)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No git repositories found")
		}
		return nil
	}

	if jsonOutput {
		return reposScanJSON(repos)
	}

	fmt.Printf("Found %d repositories\n\n", len(repos))

	// Get home for path shortening
	home, _ := os.UserHomeDir()

	// Print repos with status
	for _, repo := range repos {
		displayPath := repo.Path
		if home != "" {
			if rel, err := filepath.Rel(home, repo.Path); err == nil && !strings.HasPrefix(rel, "..") {
				displayPath = "~/" + rel
			}
		}

		var status string
		if repo.HasHooks {
			status = style.Success("[✓]")
		} else if repo.Inspection.Status.CanInstall() {
			status = style.Muted("[ ]")
		} else {
			status = style.Error("[×]") + " " + style.Warning(repo.Inspection.Status.String())
		}

		fmt.Printf("%s %s\n", status, displayPath)
	}

	// Summary
	fmt.Println()
	installed := 0
	canInstall := 0
	blocked := 0
	for _, r := range repos {
		if r.HasHooks {
			installed++
		} else if r.Inspection.Status.CanInstall() {
			canInstall++
		} else {
			blocked++
		}
	}

	fmt.Printf("Installed: %d, Available: %d", installed, canInstall)
	if blocked > 0 {
		fmt.Printf(", Blocked: %d", blocked)
	}
	fmt.Println()

	if canInstall > 0 {
		fmt.Printf("\nUse 'fp repos -i' to install hooks interactively, or 'fp setup <path>' for individual repos.\n")
	}

	return nil
}

func reposScanJSON(repos []RepoEntry) error {
	type repoJSON struct {
		Path         string `json:"path"`
		Name         string `json:"name"`
		HasHooks     bool   `json:"has_hooks"`
		CanInstall   bool   `json:"can_install"`
		Status       string `json:"status,omitempty"`
	}

	out := make([]repoJSON, 0, len(repos))
	for _, r := range repos {
		entry := repoJSON{
			Path:       r.Path,
			Name:       r.Name,
			HasHooks:   r.HasHooks,
			CanInstall: r.Inspection.Status.CanInstall(),
		}
		if !r.HasHooks && !r.Inspection.Status.CanInstall() {
			entry.Status = r.Inspection.Status.String()
		}
		out = append(out, entry)
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// CheckRepoHooks checks if fp hooks are installed in a given repo path.
func CheckRepoHooks(repoPath string) bool {
	inspection := hooks.InspectRepo(repoPath)
	return inspection.FpInstalled
}

func defaultPrintln(args ...any) (int, error) {
	return DefaultDeps().Println(args...)
}
