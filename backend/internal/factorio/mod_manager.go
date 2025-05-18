package factorio

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/snarf-dev/fsm/v2/internal/config"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

type ModInfo struct {
	Category        string       `json:"category"`
	DownloadsCount  int          `json:"downloads_count"`
	LastHighlighted string       `json:"last_highlighted_at"`
	Name            string       `json:"name"`
	Owner           string       `json:"owner"`
	Releases        []ModRelease `json:"releases"`
	Score           float64      `json:"score"`
	Summary         string       `json:"summary"`
	Thumbnail       string       `json:"thumbnail"`
	Title           string       `json:"title"`
}

type ModRelease struct {
	DownloadURL string      `json:"download_url"`
	FileName    string      `json:"file_name"`
	InfoJSON    ModInfoJSON `json:"info_json"`
	ReleasedAt  string      `json:"released_at"`
	SHA1        string      `json:"sha1"`
	Version     string      `json:"version"`
}

type ModInfoJSON struct {
	FactorioVersion string `json:"factorio_version"`
}

// DownloadMod downloads a mod.
// It optionally reports progress.
func DownloadMod(cfg *config.FSMConfig, mod string, version string) (string, error) {
	var downloadDir = filepath.Join(cfg.Factorio.Downloads, "mods")
	if downloadDir == "" {
		downloadDir = os.TempDir()
	}
	helpers.CreateDirectoryIfMissing(downloadDir)

	modInfo, err := GetModDetails(mod)
	if err != nil {
		return "", fmt.Errorf("failed to get mod details for %s: %w", mod, err)
	}
	var release *ModRelease
	for i, r := range modInfo.Releases {
		if r.Version == version {
			release = &modInfo.Releases[i]
			break
		}
	}
	if release == nil {
		return "", fmt.Errorf("version %s not found for mod %s", version, mod)
	}

	zipPath := filepath.Join(downloadDir, release.FileName)
	if _, err := os.Stat(zipPath); err == nil {
		log.Printf("Mod %s already installed", release.FileName)
	} else {
		out, err := os.Create(zipPath)
		if err != nil {
			return "", fmt.Errorf("failed to create download file: %w", err)
		}
		defer out.Close()

		downloadURL, err := createModDownloadUrl(cfg, release.DownloadURL)
		if err != nil {
			return "", err
		}
		log.Printf("Downloading %s", mod)

		resp, err := http.Get(downloadURL)
		if err != nil {
			return "", fmt.Errorf("failed to download %s: %w", downloadURL, err)
		}
		defer resp.Body.Close()

		progressWriter := &DownloadProgressWriter{
			Expected: resp.ContentLength,
			OnUpdate: func(pct int) {
				SendDownloadProgress(mod, version, "download", pct)
			},
		}
		_, err = io.Copy(io.MultiWriter(out, progressWriter), resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}

		actualSHA1, err := helpers.CalculateSHA1(zipPath)
		if err != nil {
			return "", fmt.Errorf("failed to calculate SHA1: %w", err)
		}
		if !strings.EqualFold(actualSHA1, release.SHA1) {
			return "", fmt.Errorf("SHA1 mismatch: expected %s, got %s", release.SHA1, actualSHA1)
		}
	}

	SendDownloadProgress(mod, version, "done", 100)

	log.Printf("Installed %s version %s", mod, version)

	return zipPath, nil
}

func GetModDetails(mod string) (*ModInfo, error) {
	url := fmt.Sprintf("https://mods.factorio.com/api/mods/%s", mod)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP request failed: %v\n", err)
		return nil, fmt.Errorf("failed to fetch mod details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var modInfo ModInfo
	if err := json.NewDecoder(resp.Body).Decode(&modInfo); err != nil {
		return nil, fmt.Errorf("failed to decode mod info: %w", err)
	}

	return &modInfo, nil
}

func GetAvailableMods(cfg *config.FSMConfig) ([]map[string][]string, error) {
	entries, err := os.ReadDir(filepath.Join(cfg.Factorio.Downloads, "mods"))
	if err != nil {
		return nil, err
	}

	versions := make(map[string][]string)
	for _, f := range entries {
		if f.Type().IsRegular() && strings.HasSuffix(f.Name(), ".zip") {
			filename := strings.TrimSuffix(f.Name(), ".zip")
			sepIndex := strings.LastIndex(filename, "_")
			if sepIndex == -1 {
				continue
			}
			modName := filename[:sepIndex]
			version := filename[sepIndex+1:]
			versions[modName] = append(versions[modName], version)

		}
	}
	return []map[string][]string{versions}, nil
}

func GetInstalledMods(cfg *config.FSMConfig) ([]map[string][]string, error) {
	entries, err := os.ReadDir(cfg.Factorio.ModsDir)
	if err != nil {
		return nil, err
	}

	versions := make(map[string][]string)
	for _, f := range entries {
		if f.Type().IsRegular() && strings.HasSuffix(f.Name(), ".zip") {
			filename := strings.TrimSuffix(f.Name(), ".zip")
			sepIndex := strings.LastIndex(filename, "_")
			if sepIndex == -1 {
				continue
			}
			modName := filename[:sepIndex]
			version := filename[sepIndex+1:]
			versions[modName] = append(versions[modName], version)

		}
	}
	return []map[string][]string{versions}, nil
}

func DeleteMod(cfg *config.FSMConfig, mod string, version string) error {
	targetPath := filepath.Join(cfg.Factorio.Downloads, "mods", fmt.Sprintf("%s_%s.zip", mod, version))

	if !helpers.FileExists(targetPath) {
		return fmt.Errorf("mod does not exist: %s", targetPath)
	}

	err := helpers.DeleteFile(targetPath)
	if err != nil {
		return fmt.Errorf("failed to delete mod %s-%s: %w", mod, version, err)
	}

	log.Printf("Deleted mod %s-%s\n", mod, version)
	return nil
}

func InstallMod(cfg *config.FSMConfig, mod string, version string) error {
	srcPath := filepath.Join(cfg.Factorio.Downloads, "mods", fmt.Sprintf("%s_%s.zip", mod, version))
	dstPath := filepath.Join(cfg.Factorio.ModsDir, fmt.Sprintf("%s_%s.zip", mod, version))

	if !helpers.FileExists(srcPath) {
		return fmt.Errorf("mod does not exist: %s", srcPath)
	}

	if helpers.FileExistsWildcard(filepath.Join(cfg.Factorio.ModsDir, fmt.Sprintf("%s*.zip", mod))) {
		return fmt.Errorf("mod %s is already installed", mod)
	}

	err := helpers.CopyFile(srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("failed to install mod %s-%s: %w", mod, version, err)
	}

	log.Printf("Installed mod %s-%s\n", mod, version)
	return nil
}

func UninstallMod(cfg *config.FSMConfig, mod string, version string) error {
	targetPath := filepath.Join(cfg.Factorio.ModsDir, fmt.Sprintf("%s_%s.zip", mod, version))

	if !helpers.FileExists(targetPath) {
		return fmt.Errorf("mod does not exist: %s", targetPath)
	}

	err := os.Remove(targetPath)
	if err != nil {
		return fmt.Errorf("failed to uninstall mod: %v", err)
	}

	log.Printf("Uninstalled mod %s-%s\n", mod, version)
	return nil
}

// createModDownloadUrl builds the authenticated download URL for the given mod
// using credentials from the server settings file.
func createModDownloadUrl(cfg *config.FSMConfig, uri string) (string, error) {
	username := cfg.Factorio.Username
	token := cfg.Factorio.Token

	if username == "" || token == "" {
		return "", fmt.Errorf("Unable to download without username and token")
	}

	return fmt.Sprintf("https://mods.factorio.com/%s?username=%s&token=%s", uri, username, token), nil
}
