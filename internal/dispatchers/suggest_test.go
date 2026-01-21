package dispatchers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{
			name: "identical strings",
			a:    "track",
			b:    "track",
			want: 0,
		},
		{
			name: "one character difference",
			a:    "track",
			b:    "tracky",
			want: 1,
		},
		{
			name: "typo - transposition",
			a:    "track",
			b:    "tarck",
			want: 2,
		},
		{
			name: "typo - substitution",
			a:    "status",
			b:    "stauts",
			want: 2,
		},
		{
			name: "completely different",
			a:    "track",
			b:    "xyz123",
			want: 6,
		},
		{
			name: "empty string a",
			a:    "",
			b:    "track",
			want: 5,
		},
		{
			name: "empty string b",
			a:    "track",
			b:    "",
			want: 5,
		},
		{
			name: "both empty",
			a:    "",
			b:    "",
			want: 0,
		},
		{
			name: "case insensitive",
			a:    "TRACK",
			b:    "track",
			want: 0,
		},
		{
			name: "missing letter",
			a:    "config",
			b:    "confg",
			want: 1,
		},
		{
			name: "extra letter",
			a:    "config",
			b:    "confiig",
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := levenshtein(tt.a, tt.b)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFindSimilarCommands(t *testing.T) {
	// Create a test tree with various commands
	root := &DispatchNode{
		Name:     "fp",
		Children: make(map[string]*DispatchNode),
	}

	// Add commands
	commands := []string{"track", "status", "config", "version", "check", "list", "help"}
	for _, cmd := range commands {
		root.Children[cmd] = &DispatchNode{
			Name:     cmd,
			Path:     []string{cmd},
			Children: make(map[string]*DispatchNode),
		}
	}

	tests := []struct {
		name       string
		input      string
		maxResults int
		want       []string
	}{
		{
			name:       "typo tracky suggests track",
			input:      "tracky",
			maxResults: 3,
			want:       []string{"track"},
		},
		{
			name:       "typo stauts suggests status",
			input:      "stauts",
			maxResults: 3,
			want:       []string{"status"},
		},
		{
			name:       "typo confg suggests config",
			input:      "confg",
			maxResults: 3,
			want:       []string{"config"},
		},
		{
			name:       "typo chek suggests check",
			input:      "chek",
			maxResults: 3,
			want:       []string{"check", "help"}, // both are within distance 3
		},
		{
			name:       "completely different returns nothing",
			input:      "xyz123",
			maxResults: 3,
			want:       []string{},
		},
		{
			name:       "exact match returns nothing",
			input:      "track",
			maxResults: 3,
			want:       []string{"check"}, // "check" is within distance 3 of "track"
		},
		{
			name:       "very short input returns nothing",
			input:      "c",
			maxResults: 3,
			want:       []string{}, // "c" is too far from all commands (distance > 3)
		},
		{
			name:       "nil node returns nil",
			input:      "track",
			maxResults: 3,
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := root
			if tt.name == "nil node returns nil" {
				node = nil
			}

			got := FindSimilarCommands(tt.input, node, tt.maxResults)

			if tt.want == nil {
				require.Nil(t, got)
			} else if len(tt.want) == 0 {
				require.Empty(t, got)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFindSimilarCommands_SortedByDistance(t *testing.T) {
	root := &DispatchNode{
		Name:     "fp",
		Children: make(map[string]*DispatchNode),
	}

	// Add commands at different distances from "trak"
	// trak -> track (distance 1)
	// trak -> task (distance 2)
	commands := []string{"track", "task", "version"}
	for _, cmd := range commands {
		root.Children[cmd] = &DispatchNode{
			Name:     cmd,
			Path:     []string{cmd},
			Children: make(map[string]*DispatchNode),
		}
	}

	got := FindSimilarCommands("trak", root, 3)

	// track should be first (distance 1), task second (distance 2)
	require.Len(t, got, 2)
	require.Equal(t, "track", got[0])
	require.Equal(t, "task", got[1])
}

func TestFindSimilarCommands_Subcommands(t *testing.T) {
	// Test suggestion for subcommands
	config := &DispatchNode{
		Name:     "config",
		Children: make(map[string]*DispatchNode),
	}

	subcommands := []string{"get", "set", "list"}
	for _, cmd := range subcommands {
		config.Children[cmd] = &DispatchNode{
			Name: cmd,
			Path: []string{"config", cmd},
		}
	}

	// Typo "gett" should suggest "get"
	got := FindSimilarCommands("gett", config, 3)
	require.Contains(t, got, "get")
}

func TestCollectAllCommands(t *testing.T) {
	root := &DispatchNode{
		Name:     "fp",
		Children: make(map[string]*DispatchNode),
	}

	root.Children["track"] = &DispatchNode{
		Name:     "track",
		Children: make(map[string]*DispatchNode),
	}

	config := &DispatchNode{
		Name:     "config",
		Children: make(map[string]*DispatchNode),
	}
	config.Children["get"] = &DispatchNode{Name: "get"}
	config.Children["set"] = &DispatchNode{Name: "set"}
	root.Children["config"] = config

	commands := CollectAllCommands(root, "")

	require.Contains(t, commands, "track")
	require.Contains(t, commands, "config")
	require.Contains(t, commands, "config get")
	require.Contains(t, commands, "config set")
}

func TestCollectAllCommands_NilNode(t *testing.T) {
	got := CollectAllCommands(nil, "")
	require.Nil(t, got)
}
