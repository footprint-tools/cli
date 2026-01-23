package update

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/footprint-tools/footprint-cli/internal/app"
	"github.com/footprint-tools/footprint-cli/internal/dispatchers"
)

const (
	repoOwner = "footprint-tools"
	repoName  = "footprint-cli"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Update checks for updates and installs the latest or specified version.
func Update(args []string, flags *dispatchers.ParsedFlags) error {
	return update(args, flags, NewDependencies(app.Version))
}

func update(args []string, flags *dispatchers.ParsedFlags, deps Dependencies) error {
	useTag := flags.Has("--tag")

	var targetVersion string
	if len(args) > 0 {
		targetVersion = args[0]
	}

	// If --tag flag is used, go straight to go install
	if useTag {
		if targetVersion == "" {
			return fmt.Errorf("fp: --tag requires a version argument")
		}
		return installFromSource(deps, targetVersion)
	}

	// Try to install from release
	return installFromRelease(deps, targetVersion)
}

func installFromRelease(deps Dependencies, targetVersion string) error {
	// Fetch release info
	release, err := fetchRelease(deps, targetVersion)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stdout, "No release found, trying go install...\n")
		if targetVersion == "" {
			return fmt.Errorf("fp: could not fetch latest release: %w", err)
		}
		return installFromSource(deps, targetVersion)
	}

	// Check if already up to date
	if release.TagName == deps.CurrentVersion {
		_, _ = fmt.Fprintf(deps.Stdout, "Already at latest version %s\n", release.TagName)
		return nil
	}

	// Find the right asset for this OS/arch
	assetName := fmt.Sprintf("fp_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		_, _ = fmt.Fprintf(deps.Stdout, "No binary for %s/%s, trying go install...\n", runtime.GOOS, runtime.GOARCH)
		return installFromSource(deps, release.TagName)
	}

	// Download and install
	_, _ = fmt.Fprintf(deps.Stdout, "Downloading %s...\n", release.TagName)
	if err := downloadAndInstall(deps, downloadURL); err != nil {
		return fmt.Errorf("fp: failed to install: %w", err)
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Updated to %s\n", release.TagName)
	return nil
}

func fetchRelease(deps Dependencies, version string) (*githubRelease, error) {
	var url string
	if version == "" {
		url = apiURL + "/releases/latest"
	} else {
		// Ensure version has 'v' prefix for GitHub API
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		url = apiURL + "/releases/tags/" + version
	}

	resp, err := deps.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("release not found (status %d)", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func downloadAndInstall(deps Dependencies, url string) error {
	// Get current executable path
	execPath, err := deps.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	// Download to temp file
	resp, err := deps.HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed (status %d)", resp.StatusCode)
	}

	// Create temp file for the archive
	tmpArchive, err := os.CreateTemp("", "fp-update-*.tar.gz")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpArchive.Name()) }()

	if _, err := io.Copy(tmpArchive, resp.Body); err != nil {
		return err
	}
	if err := tmpArchive.Close(); err != nil {
		return err
	}

	// Extract binary from archive
	binary, err := extractBinary(tmpArchive.Name())
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(binary) }()

	// Replace the current executable
	// First, try to remove the old one (may fail if no write permission)
	if err := os.Remove(execPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove old binary (try with sudo): %w", err)
	}

	// Copy new binary to the executable path
	if err := copyFile(binary, execPath); err != nil {
		return fmt.Errorf("could not install new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(execPath, 0755); err != nil {
		return fmt.Errorf("could not set permissions: %w", err)
	}

	return nil
}

func extractBinary(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the 'fp' binary
		if header.Typeflag == tar.TypeReg && (header.Name == "fp" || strings.HasSuffix(header.Name, "/fp")) {
			tmpBinary, err := os.CreateTemp("", "fp-binary-*")
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(tmpBinary, tr); err != nil {
				_ = tmpBinary.Close()
				_ = os.Remove(tmpBinary.Name())
				return "", err
			}
			if err := tmpBinary.Close(); err != nil {
				_ = os.Remove(tmpBinary.Name())
				return "", err
			}

			return tmpBinary.Name(), nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}

	return out.Close()
}

func installFromSource(deps Dependencies, version string) error {
	// Check if Go is available
	if err := deps.RunCommand("go", "version"); err != nil {
		return fmt.Errorf("fp: go is not installed (needed to build from source)")
	}

	// Ensure version has 'v' prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Building %s from source...\n", version)

	pkg := "github.com/" + repoOwner + "/" + repoName + "/cmd/fp@" + version
	if err := deps.RunCommand("go", "install", pkg); err != nil {
		return fmt.Errorf("fp: go install failed: %w", err)
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Installed %s (via go install)\n", version)
	_, _ = fmt.Fprintf(deps.Stdout, "Note: binary is in $GOPATH/bin or $HOME/go/bin\n")
	return nil
}
