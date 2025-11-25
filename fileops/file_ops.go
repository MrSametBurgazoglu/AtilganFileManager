package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Copy(sourcePath, destinationDir string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDirectory(sourcePath, destinationDir)
	}
	return copySingleFile(sourcePath, destinationDir)
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

		if err := Copy(sourcePath, destinationPath); err != nil {
			errors = append(errors, fmt.Errorf("error copying %s: %w", sourcePath, err))
		}
		if progress != nil {
			progress(float64(i+1) / float64(len(sourcePaths)))
		}
	}

	return errors
}

func CutFiles(sourcePaths []string, destinationDir string) []error {
	var errors []error

	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return append(errors, fmt.Errorf("failed to create destination directory %s: %w", destinationDir, err))
	}

	for _, sourcePath := range sourcePaths {
		fileName := filepath.Base(sourcePath)
		destinationPath := filepath.Join(destinationDir, fileName)

		if err := os.Rename(sourcePath, destinationPath); err != nil {
			errors = append(errors, fmt.Errorf("error cutting %s: %w", sourcePath, err))
		}
	}

	return errors
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
