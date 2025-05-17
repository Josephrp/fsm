// Package server provides logic for managing the lifecycle and configuration
// of the Factorio dedicated server instance, including log streaming,
// version tracking, and process control.
package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/snarf-dev/fsm/v2/internal/config"
	"github.com/snarf-dev/fsm/v2/internal/helpers"
)

type ServerVersion struct {
	Full    string `json:"full"`
	Branch  string `json:"branch"`
	Version string `json:"version"`
}

type ServerStatus struct {
	CanDownload  bool          `json:"can_download"`
	IsConfigured bool          `json:"is_configured"`
	Running      bool          `json:"running"`
	Version      ServerVersion `json:"version"`
}

type ServerManager struct {
	cfg            *config.FSMConfig
	cmd            *exec.Cmd
	mu             sync.Mutex
	running        bool
	logSubscribers []chan string
	Version        ServerVersion
}

// CreateManager initializes a new ServerManager, creating necessary directories
// and setting the current server version based on the configured selection.
func CreateManager(cfg *config.FSMConfig) *ServerManager {
	manager := &ServerManager{
		cfg: cfg,
	}

	manager.createFilesAndDirectories()
	manager.Version = manager.GetVersion()
	return manager
}

// Start launches the Factorio server using the configured version and options.
// It sets up log streaming and tracks the running state.
func (s *ServerManager) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	if !s.isConfigured() {
		s.InitialiseConfiguration(false)
		if !s.isConfigured() {
			return fmt.Errorf("Unable to start server until configured")
		}
	}

	s.Version = s.GetVersion()

	binaryPath := fmt.Sprintf("%s/%s/%s/factorio/bin/x64/factorio",
		s.cfg.Factorio.ServerVersions, s.cfg.Factorio.SelectedBranch, s.cfg.Factorio.SelectedVersion)
	if !helpers.FileExists(binaryPath) {
		return fmt.Errorf("%s does not exist", binaryPath)
	}
	cmd := exec.Command(binaryPath, s.buildArgs()...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go streamOutput(stdout, s)
	go streamOutput(stderr, s)

	err := cmd.Start()
	if err != nil {
		log.Printf("Error starting server: %v\n", err)
		return err
	}

	s.cmd = cmd
	s.running = true
	go func() {
		cmd.Wait()
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	log.Println("Server started")

	return nil
}

// SubscribeToLogs allows external consumers to receive log output lines from the server.
// It returns a channel that receives log lines as strings.
func (s *ServerManager) SubscribeToLogs() <-chan string {
	ch := make(chan string, 100)
	s.mu.Lock()
	s.logSubscribers = append(s.logSubscribers, ch)
	s.mu.Unlock()
	return ch
}

// Stop attempts to gracefully terminate the running Factorio server process.
func (s *ServerManager) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil {
		log.Println("Server not running")
		return nil
	}

	err := s.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Printf("Error stopping server: %v\n", err)
		return err
	}

	log.Println("Server stopped")

	s.running = false
	return nil
}

// Status returns the download availability, current running state and version of the Factorio server.
func (s *ServerManager) Status() ServerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	return ServerStatus{
		CanDownload:  s.cfg.Factorio.Username != "" && s.cfg.Factorio.Token != "",
		IsConfigured: s.isConfigured(),
		Running:      s.running,
		Version:      s.GetVersion(),
	}
}

func (s *ServerManager) InitialiseConfiguration(overwrite bool) {
	var configFiles = getConfigFiles()
	for _, f := range configFiles {
		installPath := fmt.Sprintf("%s/%s/%s/factorio/data",
			s.cfg.Factorio.ServerVersions, s.cfg.Factorio.SelectedBranch, s.cfg.Factorio.SelectedVersion)

		src := fmt.Sprintf("%s/%s.example.json", installPath, f)
		dst := fmt.Sprintf("%s/%s.json", s.cfg.Factorio.ConfigDir, f)
		if !helpers.FileExists(src) {
			log.Printf("%s not found, skipping\n", src)
			continue
		}
		if !overwrite && helpers.FileExists(dst) {
			log.Printf("%s already exists, skipping\n", dst)
			continue
		}
		err := helpers.CopyFile(src, dst)
		if err != nil {
			log.Printf("Failed to copy %s to %s: %v\n", src, dst, err)
		}
	}
}

// createFilesAndDirectories ensures that all required Factorio-related
// directories and configuration files exist.
func (s *ServerManager) createFilesAndDirectories() {
	helpers.CreateDirectoryIfMissing(s.cfg.Factorio.ConfigDir)
	helpers.CreateDirectoryIfMissing(s.cfg.Factorio.Downloads)
	helpers.CreateDirectoryIfMissing(s.cfg.Factorio.LogsDir)
	helpers.CreateDirectoryIfMissing(s.cfg.Factorio.ModsDir)
	helpers.CreateDirectoryIfMissing(s.cfg.Factorio.SavesDir)
	helpers.CreateDirectoryIfMissing(s.cfg.Factorio.ServerVersions)

	helpers.CreateFileIfMissing(fmt.Sprintf("%s/mod-list.json", s.cfg.Factorio.ModsDir), "{}")
	helpers.CreateFileIfMissing(s.cfg.Factorio.Files.AdminList, "[]")
	helpers.CreateFileIfMissing(s.cfg.Factorio.Files.BanList, "[]")
	helpers.CreateFileIfMissing(s.cfg.Factorio.Files.WhiteList, "[]")
}

// broadcastLogLine sends a log line to all subscribed log channels
// and writes the line to standard output.
func (s *ServerManager) broadcastLogLine(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.logSubscribers {
		select {
		case ch <- line:
		default:
		}
	}
	fmt.Fprintln(os.Stdout, line)
}

// buildArgs assembles the command-line arguments used to launch the Factorio server
// based on the current configuration.
func (s *ServerManager) buildArgs() []string {
	args := []string{
		"--server-settings",
		s.cfg.Factorio.Files.ServerSettings,
		"--server-adminlist",
		s.cfg.Factorio.Files.AdminList,
		"--server-banlist",
		s.cfg.Factorio.Files.BanList,
		"--server-whitelist",
		s.cfg.Factorio.Files.WhiteList,
		"--use-server-whitelist",
		"--mod-directory",
		s.cfg.Factorio.ModsDir,
		"--server-id",
		s.cfg.Factorio.Files.ServerId,
	}

	if s.cfg.Factorio.Bind != "" {
		args = append(args, "--bind", s.cfg.Factorio.Bind)
	}

	if s.cfg.Factorio.LogsDir != "" {
		args = append(args, "--console-log", fmt.Sprintf("%s/%s.log", s.cfg.Factorio.LogsDir, time.Now().Format("200601021504")))
	}

	if s.cfg.Factorio.Save == "" {
		args = append(args, "--start-server-load-latest")
	} else {
		args = append(args, "--start-server", fmt.Sprintf("%s/%s", s.cfg.Factorio.SavesDir, s.cfg.Factorio.Save))
	}

	if s.cfg.RCon.Enabled {
		if s.cfg.RCon.Bind != "" {
			args = append(args, "--rcon-bind", s.cfg.RCon.Bind)
		}
		if s.cfg.RCon.Password != "" {
			args = append(args,
				"--rcon-password", s.cfg.RCon.Password,
			)
		}
	}

	return args
}

func (s *ServerManager) isConfigured() bool {
	var configFiles = getConfigFiles()
	for _, f := range configFiles {
		var path = fmt.Sprintf("%s/%s.json", s.cfg.Factorio.ConfigDir, f)
		if !helpers.FileExists(path) {
			return false
		}
	}
	return true
}

func getConfigFiles() []string {
	var configFiles = []string{"map-gen-settings", "map-settings", "server-settings"}
	configFilesCopy := make([]string, len(configFiles))
	copy(configFilesCopy, configFiles)
	return configFilesCopy
}

// streamOutput reads from a pipe line-by-line and sends each line to the server's log stream.
func streamOutput(pipe io.ReadCloser, s *ServerManager) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		s.broadcastLogLine(line)
	}
}

// GetVersion retrieves the full version string of the selected Factorio binary.
// It executes the binary with --version and extracts the version from its output.
func (s *ServerManager) GetVersion() ServerVersion {
	if s.running {
		return s.Version
	}

	if s.cfg.Factorio.SelectedBranch == "" || s.cfg.Factorio.SelectedVersion == "" {
		return ServerVersion{}
	}

	binaryPath := fmt.Sprintf("%s/%s/%s/factorio/bin/x64/factorio",
		s.cfg.Factorio.ServerVersions, s.cfg.Factorio.SelectedBranch, s.cfg.Factorio.SelectedVersion)
	if !helpers.FileExists(binaryPath) {
		return ServerVersion{}
	}
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("failed to get Factorio version: %v\n", err)
		return ServerVersion{}
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Version: ") {
			full := strings.TrimPrefix(line, "Version: ")
			return ServerVersion{
				Full:    full,
				Branch:  s.cfg.Factorio.SelectedBranch,
				Version: s.cfg.Factorio.SelectedVersion,
			}
		}
	}
	return ServerVersion{}
}
