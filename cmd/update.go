package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

const githubRepo = "Sqoff/godot-cli"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update godot-cli to the latest version",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpdate(_ *cobra.Command, _ []string) error {
	fmt.Printf("Current version: %s\n", Version)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	release, err := fetchLatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}

	latest := release.TagName
	if latest == Version {
		fmt.Printf("Already up to date (%s)\n", Version)
		return nil
	}

	fmt.Printf("Updating %s → %s\n", Version, latest)

	assetName := platformAssetName()
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, latest)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	exePath, _ = filepath.EvalSymlinks(exePath)

	tmpPath := exePath + ".new"
	fmt.Printf("Downloading %s...\n", assetName)

	dlCtx, dlCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer dlCancel()
	if err := downloadFile(dlCtx, downloadURL, tmpPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if err := replaceBinary(exePath, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace binary failed: %w", err)
	}

	fmt.Printf("Updated to %s\n", latest)
	return nil
}

func fetchLatestRelease(ctx context.Context) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func replaceBinary(current, newPath string) error {
	bakPath := current + ".bak"
	os.Remove(bakPath)

	// Windows에서는 실행 중인 .exe를 덮어쓸 수 없지만 rename은 가능
	if err := os.Rename(current, bakPath); err != nil {
		return err
	}
	if err := os.Rename(newPath, current); err != nil {
		os.Rename(bakPath, current) // 복구
		return err
	}
	if runtime.GOOS != "windows" {
		os.Chmod(current, 0755)
	}
	os.Remove(bakPath)
	return nil
}

func platformAssetName() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("godot-cli-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
}
