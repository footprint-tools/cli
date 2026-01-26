package repo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveID(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		repoRoot  string
		want      RepoID
		wantErr   bool
	}{
		{
			name:      "SSH format - github",
			remoteURL: "git@github.com:user/repo.git",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "SSH format - gitlab",
			remoteURL: "git@gitlab.com:org/project.git",
			repoRoot:  "",
			want:      "gitlab.com/org/project",
			wantErr:   false,
		},
		{
			name:      "SSH format without .git",
			remoteURL: "git@github.com:user/repo",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "SSH format with uppercase - normalized to lowercase",
			remoteURL: "git@GitHub.com:User/Repo.git",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS format",
			remoteURL: "https://github.com/user/repo.git",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS format without .git",
			remoteURL: "https://github.com/user/repo",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS format with uppercase - normalized to lowercase",
			remoteURL: "https://GitHub.com/User/Repo.git",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "HTTP format",
			remoteURL: "http://github.com/user/repo.git",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "HTTP format without .git",
			remoteURL: "http://github.com/user/repo",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "local path when no remote",
			remoteURL: "",
			repoRoot:  "/path/to/repo",
			want:      "local:/path/to/repo",
			wantErr:   false,
		},
		{
			name:      "local path with trailing slash",
			remoteURL: "",
			repoRoot:  "/path/to/repo/",
			want:      "local:/path/to/repo",
			wantErr:   false,
		},
		{
			name:      "local path preserves case",
			remoteURL: "",
			repoRoot:  "/Path/To/MyRepo",
			want:      "local:/Path/To/MyRepo",
			wantErr:   false,
		},
		{
			name:      "remote URL preferred over local",
			remoteURL: "https://github.com/user/repo.git",
			repoRoot:  "/path/to/repo",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "invalid SSH format - missing colon",
			remoteURL: "git@github.com/user/repo.git",
			repoRoot:  "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "unsupported protocol",
			remoteURL: "ftp://example.com/repo.git",
			repoRoot:  "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "empty inputs",
			remoteURL: "",
			repoRoot:  "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "whitespace only",
			remoteURL: "   ",
			repoRoot:  "   ",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "trims whitespace from remote URL",
			remoteURL: "  https://github.com/user/repo.git  ",
			repoRoot:  "",
			want:      "github.com/user/repo",
			wantErr:   false,
		},
		{
			name:      "trims whitespace from repo root",
			remoteURL: "",
			repoRoot:  "  /path/to/repo  ",
			want:      "local:/path/to/repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeriveID(tt.remoteURL, tt.repoRoot)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDeriveID_GitProtocol(t *testing.T) {
	id, err := DeriveID("git://github.com/user/repo.git", "")
	require.NoError(t, err)
	require.Equal(t, RepoID("github.com/user/repo"), id)
}

func TestDeriveID_FileProtocol(t *testing.T) {
	id, err := DeriveID("file:///path/to/repo", "")
	require.NoError(t, err)
	require.Equal(t, RepoID("local:/path/to/repo"), id)
}

func TestDeriveID_PathTraversal(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		wantErr   bool
	}{
		{
			name:      "SSH with path traversal in host",
			remoteURL: "git@github..com:user/repo.git",
			wantErr:   true,
		},
		{
			name:      "SSH with path traversal in path",
			remoteURL: "git@github.com:../../../etc/passwd",
			wantErr:   true,
		},
		{
			name:      "HTTPS with path traversal",
			remoteURL: "https://github.com/../../../etc/passwd",
			wantErr:   true,
		},
		{
			name:      "git:// with path traversal",
			remoteURL: "git://github.com/../../../etc/passwd",
			wantErr:   true,
		},
		{
			name:      "null byte in URL",
			remoteURL: "https://github.com/user\x00/repo",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeriveID(tt.remoteURL, "")
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "path traversal")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestContainsPathTraversal(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"normal/path", false},
		{"../parent", true},
		{"path/../other", true},
		{"path/\x00null", true},
		{"safe-string", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsPathTraversal(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNewDeriver(t *testing.T) {
	d := NewDeriver()
	require.NotNil(t, d)
}

func TestDeriver_DeriveID(t *testing.T) {
	d := NewDeriver()

	// Success case
	id, err := d.DeriveID("https://github.com/user/repo.git", "")
	require.NoError(t, err)
	require.Equal(t, "github.com/user/repo", string(id))

	// Error case
	_, err = d.DeriveID("", "")
	require.Error(t, err)
}

func TestToFilesystemSafe_LeadingUnderscores(t *testing.T) {
	// Test that leading underscores are removed
	id := RepoID("//leading/slashes")
	safe := id.ToFilesystemSafe()
	require.False(t, safe[0] == '_', "should not start with underscore")
}

func TestToFilesystemSafe(t *testing.T) {
	tests := []struct {
		name string
		id   RepoID
		want string
	}{
		{
			name: "github repo",
			id:   "github.com/user/repo",
			want: "github.com__user__repo",
		},
		{
			name: "gitlab repo with organization",
			id:   "gitlab.com/org/subgroup/project",
			want: "gitlab.com__org__subgroup__project",
		},
		{
			name: "local path",
			id:   "local:/path/to/repo",
			want: "local____path__to__repo",
		},
		{
			name: "local absolute path",
			id:   "local:/Users/username/repos/myproject",
			want: "local____Users__username__repos__myproject",
		},
		{
			name: "already safe",
			id:   "simple-name",
			want: "simple-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.ToFilesystemSafe()
			require.Equal(t, tt.want, got)
		})
	}
}
