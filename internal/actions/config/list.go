package config

import (
	"encoding/json"

	"github.com/footprint-tools/cli/internal/dispatchers"
	"github.com/footprint-tools/cli/internal/domain"
	"github.com/footprint-tools/cli/internal/ui/style"
)

func List(args []string, flags *dispatchers.ParsedFlags) error {
	return list(args, flags, DefaultDeps())
}

func list(_ []string, flags *dispatchers.ParsedFlags, deps Deps) error {
	jsonOutput := flags.Has("--json")

	configMap, err := deps.GetAll()
	if err != nil {
		return err
	}

	if jsonOutput {
		return listJSON(configMap, deps)
	}

	// Show all visible keys with their values (or defaults if not set)
	for _, key := range domain.VisibleConfigKeys() {
		value, exists := configMap[key.Name]
		hasValue := exists && value != ""

		// Skip HideIfEmpty keys that have no value
		if key.HideIfEmpty && !hasValue {
			continue
		}

		if hasValue {
			_, _ = deps.Printf("%s=%s\n", style.Info(key.Name), value)
		} else if key.Default != "" {
			_, _ = deps.Printf("%s=%s %s\n", style.Info(key.Name), key.Default, style.Muted("(default)"))
		} else {
			_, _ = deps.Printf("%s= %s\n", style.Info(key.Name), style.Muted("(not set)"))
		}
	}

	return nil
}

func listJSON(configMap map[string]string, deps Deps) error {
	type configEntry struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Default  string `json:"default,omitempty"`
		IsSet    bool   `json:"is_set"`
	}

	keys := domain.VisibleConfigKeys()
	entries := make([]configEntry, 0, len(keys))
	for _, key := range keys {
		value, exists := configMap[key.Name]
		hasValue := exists && value != ""

		// Skip HideIfEmpty keys that have no value
		if key.HideIfEmpty && !hasValue {
			continue
		}

		entry := configEntry{
			Key:     key.Name,
			Default: key.Default,
			IsSet:   hasValue,
		}
		if hasValue {
			entry.Value = value
		} else {
			entry.Value = key.Default
		}
		entries = append(entries, entry)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	_, _ = deps.Println(string(data))
	return nil
}
