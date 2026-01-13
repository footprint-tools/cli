package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func RepoHooksPath(repoRoot string) (string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--git-path", "hooks")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return filepath.Clean(strings.TrimSpace(string(out))), nil
}

func GlobalHooksPath() (string, error) {
	cmd := exec.Command("git", "config", "--global", "core.hooksPath")
	out, err := cmd.Output()
	if err == nil {
		path := strings.TrimSpace(string(out))
		if path != "" {
			return path, nil
		}
	}

	homeCmd := exec.Command("git", "config", "--global", "--path", "core.hooksPath")
	homeOut, _ := homeCmd.Output()

	if strings.TrimSpace(string(homeOut)) != "" {
		return strings.TrimSpace(string(homeOut)), nil
	}

	return filepath.Join(defaultHome(), ".git", "hooks"), nil
}

func defaultHome() string {
	out, _ := exec.Command("sh", "-c", "echo $HOME").Output()
	return strings.TrimSpace(string(out))
}
