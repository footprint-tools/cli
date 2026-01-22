package repo

import (
	"errors"
	"strings"

	"github.com/Skryensya/footprint/internal/config"
)

type RepoID string

const trackedReposKey = "trackedRepos"

// containsPathTraversal checks if a string contains path traversal sequences
func containsPathTraversal(s string) bool {
	// Check for common path traversal patterns
	if strings.Contains(s, "..") {
		return true
	}
	// Check for null bytes which could be used to bypass checks
	if strings.Contains(s, "\x00") {
		return true
	}
	return false
}

func DeriveID(remoteURL, repoRoot string) (RepoID, error) {
	remoteURL = strings.TrimSpace(remoteURL)
	repoRoot = strings.TrimSpace(repoRoot)

	if remoteURL != "" {
		remoteURL = strings.TrimSuffix(remoteURL, ".git")

		if strings.HasPrefix(remoteURL, "git@") {
			parts := strings.SplitN(remoteURL, ":", 2)
			if len(parts) != 2 {
				return "", errors.New("invalid ssh remote url")
			}
			host := strings.TrimPrefix(parts[0], "git@")
			path := parts[1]

			// Validate against path traversal
			if containsPathTraversal(host) || containsPathTraversal(path) {
				return "", errors.New("invalid remote url: contains path traversal sequence")
			}

			return RepoID(host + "/" + path), nil
		}

		if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
			remoteURL = strings.TrimPrefix(remoteURL, "https://")
			remoteURL = strings.TrimPrefix(remoteURL, "http://")

			// Validate against path traversal
			if containsPathTraversal(remoteURL) {
				return "", errors.New("invalid remote url: contains path traversal sequence")
			}

			return RepoID(remoteURL), nil
		}

		// Support git:// protocol (read-only git protocol)
		if strings.HasPrefix(remoteURL, "git://") {
			remoteURL = strings.TrimPrefix(remoteURL, "git://")

			// Validate against path traversal
			if containsPathTraversal(remoteURL) {
				return "", errors.New("invalid remote url: contains path traversal sequence")
			}

			return RepoID(remoteURL), nil
		}

		// Support file:// protocol (local repositories)
		if strings.HasPrefix(remoteURL, "file://") {
			path := strings.TrimPrefix(remoteURL, "file://")
			return RepoID("local:" + path), nil
		}

		return "", errors.New("unsupported remote url format: only git@, https://, http://, git://, and file:// are supported")
	}

	if repoRoot != "" {
		clean := strings.TrimRight(repoRoot, "/")
		return RepoID("local:" + clean), nil
	}

	return "", errors.New("cannot derive repo id")
}

func ListTracked() ([]RepoID, error) {
	lines, err := config.ReadLines()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Parse(lines)
	if err != nil {
		return nil, err
	}

	value, ok := cfg[trackedReposKey]
	if !ok || strings.TrimSpace(value) == "" {
		return []RepoID{}, nil
	}

	parts := strings.Split(value, ",")
	out := make([]RepoID, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, RepoID(p))
		}
	}

	return out, nil
}

func Track(id RepoID) (bool, error) {
	lines, err := config.ReadLines()
	if err != nil {
		return false, err
	}

	current, err := ListTracked()
	if err != nil {
		return false, err
	}

	for _, existing := range current {
		if existing == id {
			return false, nil
		}
	}

	current = append(current, id)

	values := make([]string, 0, len(current))
	for _, v := range current {
		values = append(values, string(v))
	}

	lines, _ = config.Set(lines, trackedReposKey, strings.Join(values, ","))

	return true, config.WriteLines(lines)
}

func Untrack(id RepoID) (bool, error) {
	lines, err := config.ReadLines()
	if err != nil {
		return false, err
	}

	current, err := ListTracked()
	if err != nil {
		return false, err
	}

	out := make([]RepoID, 0, len(current))
	removed := false

	for _, existing := range current {
		if existing == id {
			removed = true
			continue
		}
		out = append(out, existing)
	}

	if !removed {
		return false, nil
	}

	if len(out) == 0 {
		lines, _ = config.Unset(lines, trackedReposKey)
		return true, config.WriteLines(lines)
	}

	values := make([]string, 0, len(out))
	for _, v := range out {
		values = append(values, string(v))
	}

	lines, _ = config.Set(lines, trackedReposKey, strings.Join(values, ","))

	return true, config.WriteLines(lines)
}

func IsTracked(id RepoID) (bool, error) {
	current, err := ListTracked()
	if err != nil {
		return false, err
	}

	for _, existing := range current {
		if existing == id {
			return true, nil
		}
	}

	return false, nil
}

// ToFilesystemSafe converts a RepoID to a filesystem-safe directory name.
// Transforms:
//   - "github.com/user/repo" -> "github.com__user__repo"
//   - "local:/path/to/repo" -> "local__path__to__repo"
// The transformation is deterministic and reversible (for display).
func (id RepoID) ToFilesystemSafe() string {
	idString := string(id)

	// Replace colon (from local: prefix) with double underscore
	idString = strings.ReplaceAll(idString, ":", "__")

	// Replace path separators with double underscores
	idString = strings.ReplaceAll(idString, "/", "__")

	// Remove leading underscores
	for len(idString) > 0 && idString[0] == '_' {
		idString = idString[1:]
	}

	return idString
}
