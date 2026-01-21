package git

import "github.com/Skryensya/footprint/internal/domain"

// Provider wraps git operations and implements domain.GitProvider.
type Provider struct{}

// NewProvider creates a new Git provider.
func NewProvider() *Provider {
	return &Provider{}
}

// IsAvailable checks if git is installed and accessible.
func (p *Provider) IsAvailable() bool {
	return IsAvailable()
}

// RepoRoot returns the root directory of the git repository containing the given path.
func (p *Provider) RepoRoot(path string) (string, error) {
	return RepoRoot(path)
}

// OriginURL returns the URL of the 'origin' remote.
func (p *Provider) OriginURL(repoRoot string) (string, error) {
	return OriginURL(repoRoot)
}

// ListRemotes returns a list of configured remote names.
func (p *Provider) ListRemotes(repoRoot string) ([]string, error) {
	return ListRemotes(repoRoot)
}

// GetRemoteURL returns the URL for a specific remote.
func (p *Provider) GetRemoteURL(repoRoot, remoteName string) (string, error) {
	return GetRemoteURL(repoRoot, remoteName)
}

// HeadCommit returns the current HEAD commit hash.
func (p *Provider) HeadCommit() (string, error) {
	return HeadCommit()
}

// CurrentBranch returns the current branch name.
func (p *Provider) CurrentBranch() (string, error) {
	return CurrentBranch()
}

// CommitMessage returns the most recent commit message.
func (p *Provider) CommitMessage() (string, error) {
	return CommitMessage()
}

// CommitAuthor returns the author of the most recent commit.
func (p *Provider) CommitAuthor() (string, error) {
	return CommitAuthor()
}

// Verify Provider implements domain.GitProvider
var _ domain.GitProvider = (*Provider)(nil)
