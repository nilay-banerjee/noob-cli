package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const releasesURL = "https://github.com/nilay-banerjee/noob-cli/releases"

var upgradeCmd = &cobra.Command{
	Use:          "upgrade",
	Short:        "Update noob-cli to the latest release",
	SilenceUsage: true,
	RunE:         runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	latest, err := latestReleaseVersion()
	if err != nil {
		return err
	}
	if latest == version {
		fmt.Printf("already on the latest version (%s)\n", version)
		return nil
	}
	fmt.Printf("upgrading %s -> %s\n", version, latest)

	binary, err := downloadReleaseBinary(latest)
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	staged := exe + ".new"
	if err := os.WriteFile(staged, binary, 0o755); err != nil {
		return err
	}
	if err := os.Rename(staged, exe); err != nil {
		os.Remove(staged)
		return err
	}
	fmt.Printf("updated %s\n", exe)
	return nil
}

func latestReleaseVersion() (string, error) {
	resp, err := http.Get(releasesURL + "/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tag := path.Base(resp.Request.URL.Path)
	if !strings.HasPrefix(tag, "v") {
		return "", fmt.Errorf("couldn't resolve the latest release from %s", resp.Request.URL)
	}
	return tag, nil
}

func downloadReleaseBinary(tag string) ([]byte, error) {
	url := fmt.Sprintf("%s/download/%s/noob-cli_%s_%s.tar.gz", releasesURL, tag, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s returned %s", url, resp.Status)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	archive := tar.NewReader(gz)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("noob-cli binary not found in %s", url)
		}
		if err != nil {
			return nil, err
		}
		if filepath.Base(header.Name) == "noob-cli" {
			return io.ReadAll(archive)
		}
	}
}
