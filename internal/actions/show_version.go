package actions

import (
	"encoding/json"
	"runtime"

	"github.com/footprint-tools/cli/internal/dispatchers"
)

func ShowVersion(args []string, flags *dispatchers.ParsedFlags) error {
	return showVersion(args, flags, defaultDeps())
}

func showVersion(_ []string, flags *dispatchers.ParsedFlags, deps actionDependencies) error {
	if flags.Has("--json") {
		return showVersionJSON(deps)
	}
	_, _ = deps.Printf("fp version %v\n", deps.Version())
	return nil
}

func showVersionJSON(deps actionDependencies) error {
	type versionInfo struct {
		Version  string `json:"version"`
		GoVersion string `json:"go_version"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
	}

	info := versionInfo{
		Version:   deps.Version(),
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	_, _ = deps.Println(string(data))
	return nil
}
