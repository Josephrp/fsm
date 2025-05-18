package helpers

// Package helpers provides filesystem utility functions for use throughout the application,
// such as safe creation of files and directories, and existence checks.

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// CreateFileIfMissing creates a file at the specified path if it does not exist.
// If the file is created, it writes an empty JSON object "{}" as its contents.
func CreateFileIfMissing(path string, content string) {
	_, err := os.Stat(path)
	if err == nil {
		return
	}
	if !os.IsNotExist(err) {
		log.Printf("failed to create %s, %v\n", path, err)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		log.Printf("failed to create %s, %v\n", path, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content + "\n")
	if err != nil {
		log.Printf("failed to create %s, %v\n", path, err)
	}
	log.Printf("Created %s\n", path)
}

// CopyFile copies the contents of the source file to the destination file.
// It overwrites the destination if it already exists.
// Returns an error if the copy fails at any point.
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

// DirExists returns true if the specified path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists returns true if the specified path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// CreateDirectoryIfMissing creates a directory and all necessary parents if it does not exist.
// Returns an error if directory creation fails.
func CreateDirectoryIfMissing(path string) error {
	if DirExists(path) {
		return nil
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Unable to create directory %s: %v\n", path, err)
		return err
	} else {
		log.Printf("Created directory %s\n", path)
	}

	return nil
}

// DeleteFile removes the file at the specified path if it exists.
// Logs an error if the file cannot be deleted.
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File %s does not exist, nothing to delete\n", path)
			return nil
		}
		log.Printf("Failed to delete file %s: %v\n", path, err)
		return err
	}
	log.Printf("Deleted file %s\n", path)
	return nil
}

// FileExistsWildcard returns true if any file matching the given glob pattern exists.
// For example, "file_*.zip".
func FileExistsWildcard(pattern string) bool {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Printf("Invalid glob pattern %s: %v\n", pattern, err)
		return false
	}
	return len(matches) > 0
}
