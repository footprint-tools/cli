package main

import (
	"reflect"
	"testing"
)

func TestExtractFlagsAndCommands(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantFlags    []string
		wantCommands []string
	}{
		{
			name:         "no flags or commands",
			args:         []string{},
			wantFlags:    []string{},
			wantCommands: []string{},
		},
		{
			name:         "only commands",
			args:         []string{"activity", "status"},
			wantFlags:    []string{},
			wantCommands: []string{"activity", "status"},
		},
		{
			name:         "boolean flags",
			args:         []string{"--help", "-h", "--oneline"},
			wantFlags:    []string{"--help", "-h", "--oneline"},
			wantCommands: []string{},
		},
		{
			name:         "numeric shorthand -5",
			args:         []string{"-5"},
			wantFlags:    []string{"--limit=5"},
			wantCommands: []string{},
		},
		{
			name:         "numeric shorthand -1",
			args:         []string{"-1"},
			wantFlags:    []string{"--limit=1"},
			wantCommands: []string{},
		},
		{
			name:         "numeric shorthand -10",
			args:         []string{"-10"},
			wantFlags:    []string{"--limit=10"},
			wantCommands: []string{},
		},
		{
			name:         "numeric shorthand -100",
			args:         []string{"-100"},
			wantFlags:    []string{"--limit=100"},
			wantCommands: []string{},
		},
		{
			name:         "invalid numeric -0",
			args:         []string{"-0"},
			wantFlags:    []string{"-0"},
			wantCommands: []string{},
		},
		{
			name:         "-n with space-separated value",
			args:         []string{"-n", "3"},
			wantFlags:    []string{"--limit=3"},
			wantCommands: []string{},
		},
		{
			name:         "-n with equals",
			args:         []string{"-n=5"},
			wantFlags:    []string{"-n=5"},
			wantCommands: []string{},
		},
		{
			name:         "--limit with space-separated value",
			args:         []string{"--limit", "10"},
			wantFlags:    []string{"--limit=10"},
			wantCommands: []string{},
		},
		{
			name:         "--limit with equals",
			args:         []string{"--limit=7"},
			wantFlags:    []string{"--limit=7"},
			wantCommands: []string{},
		},
		{
			name:         "mixed: command with -n value",
			args:         []string{"activity", "-n", "5"},
			wantFlags:    []string{"--limit=5"},
			wantCommands: []string{"activity"},
		},
		{
			name:         "mixed: command with numeric shorthand",
			args:         []string{"activity", "-5"},
			wantFlags:    []string{"--limit=5"},
			wantCommands: []string{"activity"},
		},
		{
			name:         "mixed: multiple flags and commands",
			args:         []string{"activity", "--oneline", "-n", "10", "--status", "pending"},
			wantFlags:    []string{"--oneline", "--limit=10", "--status=pending"},
			wantCommands: []string{"activity"},
		},
		{
			name:         "other value flags",
			args:         []string{"--status", "pending", "--source", "manual"},
			wantFlags:    []string{"--status=pending", "--source=manual"},
			wantCommands: []string{},
		},
		{
			name:         "pager flag",
			args:         []string{"--pager", "less"},
			wantFlags:    []string{"--pager=less"},
			wantCommands: []string{},
		},
		{
			name:         "-n without value",
			args:         []string{"-n"},
			wantFlags:    []string{"-n"},
			wantCommands: []string{},
		},
		{
			name:         "--limit without value",
			args:         []string{"--limit"},
			wantFlags:    []string{"--limit"},
			wantCommands: []string{},
		},
		{
			name:         "non-numeric short flag -e",
			args:         []string{"-e"},
			wantFlags:    []string{"-e"},
			wantCommands: []string{},
		},
		{
			name:         "complex real-world example",
			args:         []string{"activity", "-5", "--oneline", "--status", "pending"},
			wantFlags:    []string{"--limit=5", "--oneline", "--status=pending"},
			wantCommands: []string{"activity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlags, gotCommands := extractFlagsAndCommands(tt.args)

			if !reflect.DeepEqual(gotFlags, tt.wantFlags) {
				t.Errorf("extractFlagsAndCommands() flags = %v, want %v", gotFlags, tt.wantFlags)
			}
			if !reflect.DeepEqual(gotCommands, tt.wantCommands) {
				t.Errorf("extractFlagsAndCommands() commands = %v, want %v", gotCommands, tt.wantCommands)
			}
		})
	}
}
