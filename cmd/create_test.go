package cmd

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateChecksumFile_CreatesAndWritesChecksum(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.txt")
	data := []byte("hello checksum")

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result := createChecksumFile(filePath)
	if result.Status != Created {
		t.Fatalf("expected status %s, got %s", Created, result.Status)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	checksumPath := filePath + ".sha512"
	checksumBytes, err := os.ReadFile(checksumPath)
	if err != nil {
		t.Fatalf("read checksum file: %v", err)
	}

	hash := sha512.Sum512(data)
	expected := hex.EncodeToString(hash[:])
	if string(checksumBytes) != expected {
		t.Fatalf("checksum content mismatch: expected %q, got %q", expected, string(checksumBytes))
	}
}

func TestCreateChecksumFile_ExistingChecksumFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.txt")
	data := []byte("hello checksum")

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	checksumPath := filePath + ".sha512"
	original := []byte("existing")
	if err := os.WriteFile(checksumPath, original, 0o600); err != nil {
		t.Fatalf("write checksum file: %v", err)
	}

	result := createChecksumFile(filePath)
	if result.Status != Existing {
		t.Fatalf("expected status %s, got %s", Existing, result.Status)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	checksumBytes, err := os.ReadFile(checksumPath)
	if err != nil {
		t.Fatalf("read checksum file: %v", err)
	}

	if string(checksumBytes) != string(original) {
		t.Fatalf("checksum file should not be overwritten")
	}
}

func TestCreateChecksumFile_MissingFile(t *testing.T) {
	result := createChecksumFile(filepath.Join(t.TempDir(), "missing.txt"))
	if result.Status != Failed {
		t.Fatalf("expected status %s, got %s", Failed, result.Status)
	}
	if result.Error == nil {
		t.Fatalf("expected error, got nil")
	}
}
