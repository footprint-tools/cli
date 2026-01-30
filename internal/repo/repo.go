package repo

import (
	"errors"
	"strings"

	"github.com/footprint-tools/cli/internal/domain"
)

type RepoID string

// Deriver provides repository ID derivation.
type Deriver struct{}

// NewDeriver creates a new Deriver.
func NewDeriver() *Deriver {
	return &Deriver{}
}

// DeriveID derives a repository ID from a remote URL or local path.
func (d *Deriver) DeriveID(remoteURL, localPath string) (domain.RepoID, error) {
	id, err := DeriveID(remoteURL, localPath)
	if err != nil {
		return "", err
	}
	return domain.RepoID(id), nil
}

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

			// Normalize remote URLs to lowercase to prevent duplicates
			return RepoID(strings.ToLower(host + "/" + path)), nil
		}

		if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
			remoteURL = strings.TrimPrefix(remoteURL, "https://")
			remoteURL = strings.TrimPrefix(remoteURL, "http://")

			// Validate against path traversal
			if containsPathTraversal(remoteURL) {
				return "", errors.New("invalid remote url: contains path traversal sequence")
			}

			// Normalize remote URLs to lowercase to prevent duplicates
			return RepoID(strings.ToLower(remoteURL)), nil
		}

		// Support git:// protocol (read-only git protocol)
		if path, ok := strings.CutPrefix(remoteURL, "git://"); ok {
			if containsPathTraversal(path) {
				return "", errors.New("invalid remote url: contains path traversal sequence")
			}
			return RepoID(strings.ToLower(path)), nil
		}

		// Support file:// protocol (local repositories)
		if path, ok := strings.CutPrefix(remoteURL, "file://"); ok {
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

// ToFilesystemSafe converts a RepoID to a filesystem-safe directory name.
// Transforms:
//   - "github.com/user/repo" -> "github.com__user__repo"
//   - "local:/path/to/repo" -> "local__path__to__repo"
//
// The transformation is deterministic and reversible (for display).
func (id RepoID) ToFilesystemSafe() string {
	idString := string(id)

	// Replace colon (from local: prefix) with double underscore
	idString = strings.ReplaceAll(idString, ":", "__")

	// Replace path separators with double underscores
	idString = strings.ReplaceAll(idString, "/", "__")

	// Remove leading underscores
	return strings.TrimLeft(idString, "_")
}
