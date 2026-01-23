package repo

import (
	"os"
	"path/filepath"
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

func TestListTracked(t *testing.T) {
	tests := []struct {
		name        string
		configLines []string
		want        []RepoID
		wantErr     bool
	}{
		{
			name:        "empty config",
			configLines: []string{},
			want:        []RepoID{},
			wantErr:     false,
		},
		{
			name: "single repo",
			configLines: []string{
				"trackedRepos=github.com/user/repo",
			},
			want:    []RepoID{"github.com/user/repo"},
			wantErr: false,
		},
		{
			name: "multiple repos",
			configLines: []string{
				"trackedRepos=github.com/user/repo1,github.com/user/repo2,local:/path/to/repo",
			},
			want: []RepoID{
				"github.com/user/repo1",
				"github.com/user/repo2",
				"local:/path/to/repo",
			},
			wantErr: false,
		},
		{
			name: "repos with spaces",
			configLines: []string{
				"trackedRepos=github.com/user/repo1 , github.com/user/repo2 , local:/path/to/repo",
			},
			want: []RepoID{
				"github.com/user/repo1",
				"github.com/user/repo2",
				"local:/path/to/repo",
			},
			wantErr: false,
		},
		{
			name: "ignores other config keys",
			configLines: []string{
				"export_interval=3600",
				"trackedRepos=github.com/user/repo",
				"export_last=2024-01-01T00:00:00Z",
			},
			want:    []RepoID{"github.com/user/repo"},
			wantErr: false,
		},
		{
			name: "empty trackedRepos value",
			configLines: []string{
				"trackedRepos=",
			},
			want:    []RepoID{},
			wantErr: false,
		},
		{
			name: "whitespace-only trackedRepos value",
			configLines: []string{
				"trackedRepos=   ",
			},
			want:    []RepoID{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp HOME directory
			tempHome := t.TempDir()
			configPath := filepath.Join(tempHome, ".fprc")

			// Write config file
			if len(tt.configLines) > 0 {
				content := ""
				for _, line := range tt.configLines {
					content += line + "\n"
				}
				err := os.WriteFile(configPath, []byte(content), 0600)
				require.NoError(t, err)
			}

			// Override HOME for this test
			oldHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tempHome)
			t.Cleanup(func() {
				_ = os.Setenv("HOME", oldHome)
			})

			got, err := ListTracked()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTrack(t *testing.T) {
	tests := []struct {
		name            string
		initialRepos    []RepoID
		trackID         RepoID
		wantAdded       bool
		wantFinalRepos  []RepoID
	}{
		{
			name:           "add to empty list",
			initialRepos:   []RepoID{},
			trackID:        "github.com/user/repo",
			wantAdded:      true,
			wantFinalRepos: []RepoID{"github.com/user/repo"},
		},
		{
			name:           "add to existing list",
			initialRepos:   []RepoID{"github.com/user/repo1"},
			trackID:        "github.com/user/repo2",
			wantAdded:      true,
			wantFinalRepos: []RepoID{"github.com/user/repo1", "github.com/user/repo2"},
		},
		{
			name:           "already tracked",
			initialRepos:   []RepoID{"github.com/user/repo"},
			trackID:        "github.com/user/repo",
			wantAdded:      false,
			wantFinalRepos: []RepoID{"github.com/user/repo"},
		},
		{
			name: "already tracked among multiple",
			initialRepos: []RepoID{
				"github.com/user/repo1",
				"github.com/user/repo2",
				"github.com/user/repo3",
			},
			trackID:   "github.com/user/repo2",
			wantAdded: false,
			wantFinalRepos: []RepoID{
				"github.com/user/repo1",
				"github.com/user/repo2",
				"github.com/user/repo3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp HOME directory
			tempHome := t.TempDir()
			configPath := filepath.Join(tempHome, ".fprc")

			// Write initial config
			if len(tt.initialRepos) > 0 {
				content := "trackedRepos=" + joinRepoIDs(tt.initialRepos) + "\n"
				err := os.WriteFile(configPath, []byte(content), 0600)
				require.NoError(t, err)
			}

			// Override HOME for this test
			oldHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tempHome)
			t.Cleanup(func() {
				_ = os.Setenv("HOME", oldHome)
			})

			// Track the repo
			added, err := Track(tt.trackID)
			require.NoError(t, err)
			require.Equal(t, tt.wantAdded, added)

			// Verify final state
			got, err := ListTracked()
			require.NoError(t, err)
			require.ElementsMatch(t, tt.wantFinalRepos, got)
		})
	}
}

func TestUntrack(t *testing.T) {
	tests := []struct {
		name           string
		initialRepos   []RepoID
		untrackID      RepoID
		wantRemoved    bool
		wantFinalRepos []RepoID
	}{
		{
			name:           "remove from single-item list",
			initialRepos:   []RepoID{"github.com/user/repo"},
			untrackID:      "github.com/user/repo",
			wantRemoved:    true,
			wantFinalRepos: []RepoID{},
		},
		{
			name: "remove from multi-item list",
			initialRepos: []RepoID{
				"github.com/user/repo1",
				"github.com/user/repo2",
				"github.com/user/repo3",
			},
			untrackID:   "github.com/user/repo2",
			wantRemoved: true,
			wantFinalRepos: []RepoID{
				"github.com/user/repo1",
				"github.com/user/repo3",
			},
		},
		{
			name:           "not found in list",
			initialRepos:   []RepoID{"github.com/user/repo1"},
			untrackID:      "github.com/user/repo2",
			wantRemoved:    false,
			wantFinalRepos: []RepoID{"github.com/user/repo1"},
		},
		{
			name:           "not found in empty list",
			initialRepos:   []RepoID{},
			untrackID:      "github.com/user/repo",
			wantRemoved:    false,
			wantFinalRepos: []RepoID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp HOME directory
			tempHome := t.TempDir()
			configPath := filepath.Join(tempHome, ".fprc")

			// Write initial config
			if len(tt.initialRepos) > 0 {
				content := "trackedRepos=" + joinRepoIDs(tt.initialRepos) + "\n"
				err := os.WriteFile(configPath, []byte(content), 0600)
				require.NoError(t, err)
			}

			// Override HOME for this test
			oldHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tempHome)
			t.Cleanup(func() {
				_ = os.Setenv("HOME", oldHome)
			})

			// Untrack the repo
			removed, err := Untrack(tt.untrackID)
			require.NoError(t, err)
			require.Equal(t, tt.wantRemoved, removed)

			// Verify final state
			got, err := ListTracked()
			require.NoError(t, err)
			require.ElementsMatch(t, tt.wantFinalRepos, got)

			// If final list is empty, verify trackedRepos key was removed from config
			if len(tt.wantFinalRepos) == 0 && len(tt.initialRepos) > 0 {
				content, err := os.ReadFile(configPath)
				require.NoError(t, err)
				require.NotContains(t, string(content), "trackedRepos")
			}
		})
	}
}

func TestIsTracked(t *testing.T) {
	tests := []struct {
		name         string
		trackedRepos []RepoID
		checkID      RepoID
		want         bool
	}{
		{
			name:         "found in list",
			trackedRepos: []RepoID{"github.com/user/repo1", "github.com/user/repo2"},
			checkID:      "github.com/user/repo1",
			want:         true,
		},
		{
			name:         "not found in list",
			trackedRepos: []RepoID{"github.com/user/repo1"},
			checkID:      "github.com/user/repo2",
			want:         false,
		},
		{
			name:         "empty list",
			trackedRepos: []RepoID{},
			checkID:      "github.com/user/repo",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp HOME directory
			tempHome := t.TempDir()
			configPath := filepath.Join(tempHome, ".fprc")

			// Write initial config
			if len(tt.trackedRepos) > 0 {
				content := "trackedRepos=" + joinRepoIDs(tt.trackedRepos) + "\n"
				err := os.WriteFile(configPath, []byte(content), 0600)
				require.NoError(t, err)
			}

			// Override HOME for this test
			oldHome := os.Getenv("HOME")
			_ = os.Setenv("HOME", tempHome)
			t.Cleanup(func() {
				_ = os.Setenv("HOME", oldHome)
			})

			got, err := IsTracked(tt.checkID)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
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

// Helper function to join RepoIDs for config file
func joinRepoIDs(ids []RepoID) string {
	if len(ids) == 0 {
		return ""
	}
	result := string(ids[0])
	for i := 1; i < len(ids); i++ {
		result += "," + string(ids[i])
	}
	return result
}
