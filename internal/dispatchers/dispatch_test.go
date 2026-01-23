package dispatchers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Mock action functions for testing
func mockAction(args []string, flags *ParsedFlags) error {
	return nil
}

// Helper to create a simple tree for testing
func createTestTree() *DispatchNode {
	root := &DispatchNode{
		Name:     "fp",
		Path:     []string{},
		Summary:  "Test CLI",
		Children: make(map[string]*DispatchNode),
		Flags: []FlagDescriptor{
			{Names: []string{"--verbose", "-v"}, Description: "Verbose output"},
			{Names: []string{"--help", "-h"}, Description: "Show help"},
		},
	}

	// Add a simple command
	version := &DispatchNode{
		Name:     "version",
		Path:     []string{"version"},
		Summary:  "Show version",
		Action:   mockAction,
		Children: make(map[string]*DispatchNode),
	}
	root.Children["version"] = version

	// Add a command with args
	track := &DispatchNode{
		Name:    "track",
		Path:    []string{"track"},
		Summary: "Track a repository",
		Args: []ArgSpec{
			{Name: "path", Description: "Repository path", Required: true},
		},
		Action:   mockAction,
		Children: make(map[string]*DispatchNode),
		Flags: []FlagDescriptor{
			{Names: []string{"--remote"}, Description: "Remote name", ValueHint: "name"},
		},
	}
	root.Children["track"] = track

	// Add a group with subcommands
	config := &DispatchNode{
		Name:     "config",
		Path:     []string{"config"},
		Summary:  "Manage configuration",
		Children: make(map[string]*DispatchNode),
	}
	root.Children["config"] = config

	configSet := &DispatchNode{
		Name:    "set",
		Path:    []string{"config", "set"},
		Summary: "Set config value",
		Args: []ArgSpec{
			{Name: "key", Description: "Config key", Required: true},
			{Name: "value", Description: "Config value", Required: true},
		},
		Action:   mockAction,
		Children: make(map[string]*DispatchNode),
	}
	config.Children["set"] = configSet

	configGet := &DispatchNode{
		Name:    "get",
		Path:    []string{"config", "get"},
		Summary: "Get config value",
		Args: []ArgSpec{
			{Name: "key", Description: "Config key", Required: true},
		},
		Action:   mockAction,
		Children: make(map[string]*DispatchNode),
	}
	config.Children["get"] = configGet

	return root
}

func TestDispatch_SimpleCommand(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	res, err := Dispatch(root, []string{"version"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Node)
	require.Equal(t, "version", res.Node.Name)
	require.NotNil(t, res.Execute)
	require.Empty(t, res.Args)
}

func TestDispatch_CommandWithArgs(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	res, err := Dispatch(root, []string{"track", "/path/to/repo"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Node)
	require.Equal(t, "track", res.Node.Name)
	require.Equal(t, []string{"/path/to/repo"}, res.Args)
	require.NotNil(t, res.Execute)
}

func TestDispatch_NestedCommand(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	res, err := Dispatch(root, []string{"config", "set", "key1", "value1"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Node)
	require.Equal(t, "set", res.Node.Name)
	require.Equal(t, []string{"key1", "value1"}, res.Args)
}

func TestDispatch_MissingRequiredArg(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	_, err := Dispatch(root, []string{"track"}, flags)
	require.Error(t, err)
	require.Contains(t, err.Error(), "path")
}

func TestDispatch_HelpFlag(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name   string
		tokens []string
		flags  []string
	}{
		{
			name:   "--help flag on root",
			tokens: []string{},
			flags:  []string{"--help"},
		},
		{
			name:   "-h flag on root",
			tokens: []string{},
			flags:  []string{"-h"},
		},
		{
			name:   "--help flag on command",
			tokens: []string{"version"},
			flags:  []string{"--help"},
		},
		{
			name:   "--help flag on nested command",
			tokens: []string{"config", "set"},
			flags:  []string{"--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := NewParsedFlags(tt.flags)
			res, err := Dispatch(root, tt.tokens, flags)
			require.NoError(t, err)
			require.NotNil(t, res.Execute)
			// Should be help action (we can't directly test the function, but it should be set)
		})
	}
}

func TestDispatch_HelpCommand(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	tests := []struct {
		name   string
		tokens []string
	}{
		{
			name:   "help for root",
			tokens: []string{"help"},
		},
		{
			name:   "help for version",
			tokens: []string{"help", "version"},
		},
		{
			name:   "help for config",
			tokens: []string{"help", "config"},
		},
		{
			name:   "help for nested command",
			tokens: []string{"help", "config", "set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Dispatch(root, tt.tokens, flags)
			require.NoError(t, err)
			require.NotNil(t, res.Execute)
		})
	}
}

func TestDispatch_InvalidCommand(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	_, err := Dispatch(root, []string{"help", "nonexistent"}, flags)
	require.Error(t, err)
}

func TestDispatch_NoCommandShowsHelp(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	res, err := Dispatch(root, []string{}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Execute)
	require.Equal(t, 1, res.ExitCode, "should exit with code 1 when no command")
}

func TestDispatch_GroupWithoutSubcommandShowsHelp(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	res, err := Dispatch(root, []string{"config"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Execute)
	require.Equal(t, 0, res.ExitCode, "should exit with code 0 for intermediate nodes")
}

func TestDispatch_InvalidFlag(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{"--invalid-flag"})

	_, err := Dispatch(root, []string{"version"}, flags)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid-flag")
}

func TestDispatch_ValidLocalFlag(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{"--remote=origin"})

	res, err := Dispatch(root, []string{"track", "/path"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Node)
	require.Equal(t, "track", res.Node.Name)
}

func TestDispatch_ValidGlobalFlag(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{"--verbose"})

	res, err := Dispatch(root, []string{"version"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Node)
}

func TestHasHelpFlag(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  bool
	}{
		{
			name:  "no help flag",
			flags: []string{"--verbose"},
			want:  false,
		},
		{
			name:  "has --help",
			flags: []string{"--help"},
			want:  true,
		},
		{
			name:  "has -h",
			flags: []string{"-h"},
			want:  true,
		},
		{
			name:  "has both",
			flags: []string{"--help", "-h"},
			want:  true,
		},
		{
			name:  "has --help among others",
			flags: []string{"--verbose", "--help", "--debug"},
			want:  true,
		},
		{
			name:  "empty flags",
			flags: []string{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := NewParsedFlags(tt.flags)
			got := hasHelpFlag(flags)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidFlagsForNode(t *testing.T) {
	root := &DispatchNode{
		Name: "root",
		Flags: []FlagDescriptor{
			{Names: []string{"--verbose", "-v"}},
			{Names: []string{"--quiet", "-q"}},
		},
	}

	node := &DispatchNode{
		Name: "command",
		Flags: []FlagDescriptor{
			{Names: []string{"--local"}},
			{Names: []string{"--force", "-f"}},
		},
	}

	valid := validFlagsForNode(node, root)

	// Should include both root and node flags
	require.True(t, valid["--verbose"])
	require.True(t, valid["-v"])
	require.True(t, valid["--quiet"])
	require.True(t, valid["-q"])
	require.True(t, valid["--local"])
	require.True(t, valid["--force"])
	require.True(t, valid["-f"])

	// Should not include random flags
	require.False(t, valid["--random"])
}

func TestValidateFlags(t *testing.T) {
	valid := map[string]bool{
		"--verbose": true,
		"-v":        true,
		"--output":  true,
	}

	tests := []struct {
		name    string
		flags   []string
		wantErr bool
	}{
		{
			name:    "all valid flags",
			flags:   []string{"--verbose", "--output=file.txt"},
			wantErr: false,
		},
		{
			name:    "valid short flag",
			flags:   []string{"-v"},
			wantErr: false,
		},
		{
			name:    "empty flags",
			flags:   []string{},
			wantErr: false,
		},
		{
			name:    "invalid flag",
			flags:   []string{"--invalid"},
			wantErr: true,
		},
		{
			name:    "valid and invalid mixed",
			flags:   []string{"--verbose", "--invalid"},
			wantErr: true,
		},
		{
			name:    "invalid flag with value",
			flags:   []string{"--invalid=value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := NewParsedFlags(tt.flags)
			err := validateFlags(flags, valid)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name    string
		spec    []ArgSpec
		args    []string
		wantErr bool
	}{
		{
			name: "all required args provided",
			spec: []ArgSpec{
				{Name: "arg1", Required: true},
				{Name: "arg2", Required: true},
			},
			args:    []string{"value1", "value2"},
			wantErr: false,
		},
		{
			name: "missing required arg",
			spec: []ArgSpec{
				{Name: "arg1", Required: true},
				{Name: "arg2", Required: true},
			},
			args:    []string{"value1"},
			wantErr: true,
		},
		{
			name: "optional arg not provided",
			spec: []ArgSpec{
				{Name: "arg1", Required: true},
				{Name: "arg2", Required: false},
			},
			args:    []string{"value1"},
			wantErr: false,
		},
		{
			name:    "no args required, none provided",
			spec:    []ArgSpec{},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "extra args provided",
			spec: []ArgSpec{
				{Name: "arg1", Required: true},
			},
			args:    []string{"value1", "value2", "value3"},
			wantErr: false,
		},
		{
			name: "no required args",
			spec: []ArgSpec{
				{Name: "arg1", Required: false},
				{Name: "arg2", Required: false},
			},
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.spec, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResolveNode(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name     string
		path     []string
		wantNode *DispatchNode
		wantNil  bool
	}{
		{
			name:     "resolve root",
			path:     []string{},
			wantNode: root,
			wantNil:  false,
		},
		{
			name:     "resolve direct child",
			path:     []string{"version"},
			wantNode: root.Children["version"],
			wantNil:  false,
		},
		{
			name:     "resolve nested child",
			path:     []string{"config", "set"},
			wantNode: root.Children["config"].Children["set"],
			wantNil:  false,
		},
		{
			name:    "resolve non-existent path",
			path:    []string{"nonexistent"},
			wantNil: true,
		},
		{
			name:    "resolve partially non-existent path",
			path:    []string{"config", "nonexistent"},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveNode(root, tt.path)
			if tt.wantNil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, tt.wantNode, got)
			}
		})
	}
}

func TestDispatch_HelpTopics(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	res, err := Dispatch(root, []string{"help", "topics"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Execute)
	require.Equal(t, root, res.Node)
}

func TestDispatch_HelpTopic(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	// "overview" is a valid help topic
	res, err := Dispatch(root, []string{"help", "overview"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Execute)
}

func TestDispatch_HelpBeforeCommand(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	// "help" before a valid command path
	res, err := Dispatch(root, []string{"help", "config", "set"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Execute)
	require.Equal(t, "set", res.Node.Name)
}

func TestDispatch_CommandThenHelp(t *testing.T) {
	root := createTestTree()
	flags := NewParsedFlags([]string{})

	// "config help" pattern - help after a group shows help for the group
	res, err := Dispatch(root, []string{"config", "help"}, flags)
	require.NoError(t, err)
	require.NotNil(t, res.Execute)
}
