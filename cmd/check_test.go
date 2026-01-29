package cmd

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckChecksumFile_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.txt")

	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result := checkChecksumFile(filePath)
	if result.Status != NotFound {
		t.Fatalf("expected status %s, got %s", NotFound, result.Status)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
}

func TestCheckChecksumFile_MatchCaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.txt")
	data := []byte("hello")

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hash := sha512.Sum512(data)
	checksum := strings.ToUpper(hex.EncodeToString(hash[:]))
	if err := os.WriteFile(filePath+".sha512", []byte(checksum), 0o600); err != nil {
		t.Fatalf("write checksum file: %v", err)
	}

	result := checkChecksumFile(filePath)
	if result.Status != Match {
		t.Fatalf("expected status %s, got %s", Match, result.Status)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
}

func TestCheckChecksumFile_NotMatch(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "data.txt")

	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := os.WriteFile(filePath+".sha512", []byte("deadbeef"), 0o600); err != nil {
		t.Fatalf("write checksum file: %v", err)
	}

	result := checkChecksumFile(filePath)
	if result.Status != NotMatch {
		t.Fatalf("expected status %s, got %s", NotMatch, result.Status)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
}

func TestCheckChecksumFile_MissingFile(t *testing.T) {
	result := checkChecksumFile(filepath.Join(t.TempDir(), "missing.txt"))
	if result.Status != CheckingFailed {
		t.Fatalf("expected status %s, got %s", CheckingFailed, result.Status)
	}
	if result.Error == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestCheckChecksumFile_LockedFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not enforced on Windows")
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "locked.txt")

	if err := os.WriteFile(filePath, []byte("secret"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := os.Chmod(filePath, 0o000); err != nil {
		t.Fatalf("chmod file: %v", err)
	}
	defer func() {
		_ = os.Chmod(filePath, 0o600)
	}()

	if f, err := os.Open(filePath); err == nil {
		_ = f.Close()
		t.Skip("unable to enforce read permissions in this environment")
	}

	result := checkChecksumFile(filePath)
	if result.Status != LockedVerification {
		t.Fatalf("expected status %s, got %s", LockedVerification, result.Status)
	}
	if result.Error == nil {
		t.Fatalf("expected error, got nil")
	}
}
