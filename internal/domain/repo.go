package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/footprint-tools/cli/internal/usage"
)

// RepoID constants for filesystem-safe conversion.
const (
	RepoIDMaxLength   = 100 // Maximum length before hashing
	RepoIDTruncateLen = 50  // Length to truncate to when hashing
	LocalRepoIDPrefix = "local/"
)

// RepoID represents a unique identifier for a repository.
// It is derived from the remote URL or local path.
type RepoID string

// String returns the string representation of the RepoID.
func (id RepoID) String() string {
	return string(id)
}

// IsEmpty returns true if the RepoID is empty.
func (id RepoID) IsEmpty() bool {
	return id == ""
}

// ToFilesystemSafe converts the RepoID to a filesystem-safe string.
// Replaces special characters with underscores and limits length.
func (id RepoID) ToFilesystemSafe() string {
	s := string(id)

	// Replace common separators with underscores
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "@", "_")

	// Remove any remaining unsafe characters
	re := regexp.MustCompile(`[^a-zA-Z0-9_.-]`)
	s = re.ReplaceAllString(s, "_")

	// Collapse multiple underscores
	re = regexp.MustCompile(`_+`)
	s = re.ReplaceAllString(s, "_")

	// Trim leading/trailing underscores
	s = strings.Trim(s, "_")

	// If too long, hash it
	if len(s) > RepoIDMaxLength {
		hash := sha256.Sum256([]byte(id))
		s = s[:RepoIDTruncateLen] + "_" + hex.EncodeToString(hash[:8])
	}

	return s
}

// DeriveRepoID derives a RepoID from a remote URL or local path.
// If remoteURL is provided, it takes precedence over the local path.
func DeriveRepoID(remoteURL, localPath string) (RepoID, error) {
	if remoteURL != "" {
		return deriveFromRemote(remoteURL)
	}
	return deriveFromPath(localPath)
}

// deriveFromRemote extracts a normalized repository identifier from a remote URL.
func deriveFromRemote(remoteURL string) (RepoID, error) {
	url := remoteURL

	// Remove protocol prefix
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git://")

	// Handle SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
	}

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Remove trailing slashes
	url = strings.TrimSuffix(url, "/")

	if url == "" {
		return "", usage.InvalidRepo()
	}

	return RepoID(url), nil
}

// deriveFromPath creates a local identifier from the repository path.
func deriveFromPath(localPath string) (RepoID, error) {
	if localPath == "" {
		return "", usage.InvalidPath()
	}

	// Clean and get absolute-like representation
	path := filepath.Clean(localPath)

	// Use "local/" prefix to distinguish from remote repos
	return RepoID(LocalRepoIDPrefix + path), nil
}
