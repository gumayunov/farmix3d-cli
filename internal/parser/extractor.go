package parser

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ExtractArchive(archivePath string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open 3MF archive: %w", err)
	}
	defer reader.Close()

	tempDir, err := os.MkdirTemp("/tmp", "3mf_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	for _, file := range reader.File {
		if err := extractFile(file, tempDir); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	return tempDir, nil
}

func extractFile(file *zip.File, destDir string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	destPath := filepath.Join(destDir, file.Name)

	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.FileInfo().Mode())
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

func CleanupTemp(tempDir string) error {
	return os.RemoveAll(tempDir)
}