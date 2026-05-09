package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Copy(sourcePath, destinationPath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source path %s: %w", sourcePath, err)
	}

	if info.IsDir() {
		return copyDirectory(sourcePath, destinationPath)
	}
	return copySingleFile(sourcePath, destinationPath)
}

func copyDirectory(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destination, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(sourcePath, destinationPath); err != nil {
				return err
			}
		} else {
			if err := copySingleFile(sourcePath, destinationPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFiles(sourcePaths []string, destinationDir string, progress func(float64)) []error {
	var errors []error

	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return append(errors, fmt.Errorf("failed to create destination directory %s: %w", destinationDir, err))
	}

	for i, sourcePath := range sourcePaths {
		baseName := filepath.Base(sourcePath)
		destinationPath := filepath.Join(destinationDir, baseName)
		
		// Conflict resolution: rename if destination exists
		if _, err := os.Stat(destinationPath); err == nil {
			destinationPath = GetUniquePath(destinationPath)
		}

		if err := Copy(sourcePath, destinationPath); err != nil {
			errors = append(errors, fmt.Errorf("error copying %s: %w", sourcePath, err))
		}
		if progress != nil {
			progress(float64(i+1) / float64(len(sourcePaths)))
		}
	}

	return errors
}

func CutFiles(sourcePaths []string, destinationDir string, progress func(float64)) []error {
	var errors []error

	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return append(errors, fmt.Errorf("failed to create destination directory %s: %w", destinationDir, err))
	}

	for i, sourcePath := range sourcePaths {
		fileName := filepath.Base(sourcePath)
		destinationPath := filepath.Join(destinationDir, fileName)

		// Conflict resolution: rename if destination exists
		if _, err := os.Stat(destinationPath); err == nil {
			destinationPath = GetUniquePath(destinationPath)
		}

		if err := os.Rename(sourcePath, destinationPath); err != nil {
			errors = append(errors, fmt.Errorf("error cutting %s: %w", sourcePath, err))
		}
		if progress != nil {
			progress(float64(i+1) / float64(len(sourcePaths)))
		}
	}

	return errors
}

func GetUniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		if i > 1000 { // Safety break
			return path
		}
	}
}

func copySingleFile(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	stat, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return destinationFile.Sync()
}
