package git

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type DiffStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
	Files        []FileStat
}

type FileStat struct {
	Path       string
	Insertions int
	Deletions  int
}

func IsAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func RepoRoot(path string) (string, error) {
	return runGit("-C", path, "rev-parse", "--show-toplevel")
}

func OriginURL(repoRoot string) (string, error) {
	return runGit("-C", repoRoot, "remote", "get-url", "origin")
}

func HeadCommit() (string, error) {
	return runGit("rev-parse", "HEAD")
}

func CommitMessage() (string, error) {
	out, err := exec.Command(
		"git", "show", "-s", "--format=%s",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func CurrentBranch() (string, error) {
	return runGit("rev-parse", "--abbrev-ref", "HEAD")
}

func CommitAuthor() (string, error) {
	return runGit("show", "-s", "--format=%an <%ae>", "HEAD")
}

func CommitSubject() (string, error) {
	return runGit("show", "-s", "--format=%s", "HEAD")
}

func CommitDiffStats() (DiffStats, error) {
	out, err := runGit(
		"diff-tree",
		"--no-commit-id",
		"--numstat",
		"-r",
		"HEAD",
	)
	if err != nil {
		return DiffStats{}, err
	}

	stats := DiffStats{}
	lines := strings.Split(strings.TrimSpace(out), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}

		ins := parseNumstat(parts[0])
		del := parseNumstat(parts[1])
		path := parts[2]

		stats.FilesChanged++
		stats.Insertions += ins
		stats.Deletions += del

		stats.Files = append(stats.Files, FileStat{
			Path:       path,
			Insertions: ins,
			Deletions:  del,
		})
	}

	return stats, nil
}

func parseNumstat(v string) int {
	if v == "-" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
