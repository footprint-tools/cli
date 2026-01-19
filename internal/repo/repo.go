package repo

import (
	"errors"
	"strings"

	"github.com/Skryensya/footprint/internal/config"
)

type RepoID string

const trackedReposKey = "trackedRepos"

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
			return RepoID(host + "/" + parts[1]), nil
		}

		if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
			remoteURL = strings.TrimPrefix(remoteURL, "https://")
			remoteURL = strings.TrimPrefix(remoteURL, "http://")
			return RepoID(remoteURL), nil
		}

		return "", errors.New("unsupported remote url format")
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
	s := string(id)

	// Replace colon (from local: prefix) with double underscore
	s = strings.ReplaceAll(s, ":", "__")

	// Replace path separators with double underscores
	s = strings.ReplaceAll(s, "/", "__")

	// Remove leading underscores
	for len(s) > 0 && s[0] == '_' {
		s = s[1:]
	}

	return s
}
