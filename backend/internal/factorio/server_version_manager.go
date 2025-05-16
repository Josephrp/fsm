package factorio

// Package factorio provides logic for managing Factorio server versions,
// including downloading, extracting, tracking progress, selecting, and uninstalling.

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/snarf-dev/fsm/v2/internal/config"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
	"github.com/ulikunitz/xz"
)

type DownloadProgressWriter struct {
	Total    int64
	Expected int64
	OnUpdate func(percent int)
	lastEmit int
}

type stageProgress struct {
	Stage   string `json:"stage"`
	Percent int    `json:"percent"`
}

var (
	downloadSubscribersMu sync.Mutex
	downloadSubscribers   = make(map[string][]chan stageProgress)
)

// GetInstalledFactorioVersions scans the given path for installed server versions
// organized by branch and returns a map of branch to version names.
func GetInstalledFactorioVersions(path string) (map[string][]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	versions := make(map[string][]string)
	for _, branch := range entries {
		if branch.IsDir() {
			branchPath := filepath.Join(path, branch.Name())
			subentries, err := os.ReadDir(branchPath)
			if err != nil {
				continue
			}
			for _, version := range subentries {
				if version.IsDir() {
					versions[branch.Name()] = append(versions[branch.Name()], version.Name())
				}
			}
		}
	}
	return versions, nil
}

// DownloadAndExtractVersion downloads and extracts the specified Factorio version.
// It optionally reports progress and reuses the download if it already exists.
func DownloadAndExtractVersion(cfg *config.FSMConfig, branch string, version string) (string, error) {
	var downloadDir = cfg.Factorio.Downloads
	if downloadDir == "" {
		downloadDir = os.TempDir()
	}

	targetDir := filepath.Join(downloadDir, branch)
	helpers.CreateDirectoryIfMissing(targetDir)

	tarPath := filepath.Join(targetDir, fmt.Sprintf("factorio-headless_linux_%s.tar.xz", version))
	if _, err := os.Stat(tarPath); err == nil {
		log.Printf("Using cached download for version %s", version)
	} else {
		out, err := os.Create(tarPath)
		if err != nil {
			return "", fmt.Errorf("failed to create download file: %w", err)
		}
		defer out.Close()

		downloadURL, err := createDownloadUrl(cfg, version)
		if err != nil {
			return "", err
		}
		log.Printf("Downloading server version %s", version)

		resp, err := http.Get(downloadURL)
		if err != nil {
			return "", fmt.Errorf("failed to download %s: %w", downloadURL, err)
		}
		defer resp.Body.Close()

		progressWriter := &DownloadProgressWriter{
			Expected: resp.ContentLength,
			OnUpdate: func(pct int) {
				SendDownloadProgress(branch, version, "download", pct)
			},
		}
		_, err = io.Copy(io.MultiWriter(out, progressWriter), resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}
	}

	targetPath := filepath.Join(cfg.Factorio.ServerVersions, branch, version)
	err := helpers.CreateDirectoryIfMissing(targetPath)
	if err != nil {
		return "", fmt.Errorf("Unable to create directory %s: %w\n", targetPath, err)
	}

	log.Printf("Extracting server %s version %s", branch, version)
	SendDownloadProgress(branch, version, "unpack", 0)
	err = extractTarXz(tarPath, targetPath, func(pct int) {
		SendDownloadProgress(branch, version, "unpack", pct)
	})
	if err != nil {
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	SendDownloadProgress(branch, version, "done", 100)

	log.Printf("Installed server %s version %s", branch, version)

	return targetPath, nil
}

// SubscribeDownloadProgress registers a listener for progress updates during
// download and unpack stages for a specific branch and version.
func SubscribeDownloadProgress(branch string, version string) <-chan stageProgress {
	downloadSubscribersMu.Lock()
	defer downloadSubscribersMu.Unlock()

	ch := make(chan stageProgress, 100)
	key := fmt.Sprintf("%s-%s", branch, version)
	downloadSubscribers[key] = append(downloadSubscribers[key], ch)
	return ch
}

// SendDownloadProgress emits progress updates to all subscribers for a given
// branch/version and stage (e.g., "download", "unpack").
func SendDownloadProgress(branch string, version string, stage string, percent int) {
	downloadSubscribersMu.Lock()
	defer downloadSubscribersMu.Unlock()

	key := fmt.Sprintf("%s-%s", branch, version)
	subs := downloadSubscribers[key]
	active := subs[:0]
	for _, ch := range subs {
		select {
		case ch <- stageProgress{Stage: stage, Percent: percent}:
			active = append(active, ch)
		default:
			// Drop if blocked
		}
	}
	downloadSubscribers[key] = active
}

// SelectVersion updates the configuration to use a specific branch and version
// and persists the selection to the config file.
func SelectVersion(cfg *config.FSMConfig, branch string, version string) error {
	targetPath := filepath.Join(cfg.Factorio.ServerVersions, branch, version)

	if !helpers.DirExists(targetPath) {
		return fmt.Errorf("version directory does not exist: %s", targetPath)
	}

	cfg.Factorio.SelectedBranch = branch
	cfg.Factorio.SelectedVersion = version
	err := cfg.SaveToFile()
	if err != nil {
		return err
	}

	log.Printf("Selected server %s version %s\n", branch, version)

	return nil
}

// UninstallVersion removes the server files for a given branch and version.
func UninstallVersion(cfg *config.FSMConfig, branch string, version string) error {
	targetPath := filepath.Join(cfg.Factorio.ServerVersions, branch, version)

	if !helpers.DirExists(targetPath) {
		return fmt.Errorf("version directory does not exist: %s", targetPath)
	}

	err := os.RemoveAll(targetPath)
	if err != nil {
		return fmt.Errorf("failed to remove version directory: %w", err)
	}

	log.Printf("Uninstalled server version %s from %s\n", version, targetPath)
	return nil
}

func (pw *DownloadProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Total += int64(n)
	if pw.Expected > 0 {
		percent := int(float64(pw.Total) * 100 / float64(pw.Expected))
		if percent > pw.lastEmit {
			pw.lastEmit = percent
			if pw.OnUpdate != nil {
				pw.OnUpdate(percent)
			}
		}
	}
	return n, nil
}

// createDownloadUrl builds the authenticated download URL for the given version
// using credentials from the server settings file.
func createDownloadUrl(cfg *config.FSMConfig, version string) (string, error) {
	username := cfg.Factorio.Username
	token := cfg.Factorio.Token

	if username == "" || token == "" {
		return "", fmt.Errorf("Unable to download without username and token")
	}

	return fmt.Sprintf("https://www.factorio.com/get-download/%s/headless/linux64?username=%s&token=%s", version, username, token), nil
}

// extractTarXz decompresses a .tar.xz archive to the target directory.
// It optionally reports progress as the file is written and unpacked.
func extractTarXz(archivePath, targetDir string, onUpdate func(int)) error {
	tarPath := archivePath[:len(archivePath)-3]
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create intermediate tar file: %w", err)
	}
	defer tarFile.Close()

	xzFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open xz archive: %w", err)
	}
	defer xzFile.Close()

	xzReader, err := xz.NewReader(xzFile)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	expectedSize := getUncompressedSizeXZ(archivePath)
	var totalWritten int64
	buffer := make([]byte, 32*1024)
	for {
		n, err := xzReader.Read(buffer)
		if n > 0 {
			if _, werr := tarFile.Write(buffer[:n]); werr != nil {
				return fmt.Errorf("failed writing tar: %w", werr)
			}
			totalWritten += int64(n)
			if onUpdate != nil && expectedSize > 0 {
				pct := int(float64(totalWritten) * 100 / float64(expectedSize))
				if pct > 100 {
					pct = 100
				}
				onUpdate(pct / 2)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error decompressing xz: %w", err)
		}
	}

	tarFileRead, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to reopen tar file: %w", err)
	}
	defer tarFileRead.Close()

	if _, err := tarFileRead.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek tar: %w", err)
	}
	tarReader := tar.NewReader(tarFileRead)

	i := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %w", err)
		}

		destPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("mkdir failed: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("create file failed: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("copy file failed: %w", err)
			}
			outFile.Close()
			if err := os.Chmod(destPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("chmod failed: %w", err)
			}
		}

		if onUpdate != nil {
			pct := 50 + (i % 50)
			if pct > 100 {
				pct = 100
			}
			onUpdate(pct)
		}
		i++
	}

	_ = os.Remove(tarPath)
	return nil
}

// getUncompressedSizeXZ parses the uncompressed size of a .xz file using the `xz --list` tool.
func getUncompressedSizeXZ(path string) int64 {
	cmd := exec.Command("xz", "--list", path)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("failed to run xz --list: %v", err)
		return 0
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) >= 5 {
		unit := strings.TrimSpace(fields[5])
		val := strings.TrimSpace(fields[4])
		switch {
		case unit == "MiB":
			if size, err := strconv.ParseFloat(val, 64); err == nil {
				return int64(size * 1024 * 1024)
			}
		case unit == "GiB":
			if size, err := strconv.ParseFloat(val, 64); err == nil {
				return int64(size * 1024 * 1024 * 1024)
			}
		}
	}

	return 0
}
