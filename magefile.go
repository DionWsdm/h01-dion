//go:build mage

package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
var Default = Build

// binaries maps target binary names to their Go package source paths
var binaries = map[string]string{
	"tcp-server":  "./samples/tcp/server",
	"tcp-client":  "./samples/tcp/client",
	"udp-server":  "./samples/udp/server",
	"udp-client":  "./samples/udp/client",
	"quic-server": "./samples/quic/server",
	"quic-client": "./samples/quic/client",
	"codec":       "./samples/codec",
	"publisher":   "./project/publisher",
	"subscriber":  "./project/subscriber",
}

// Build compiles all sample and project executables into the bin/ directory
func Build() error {
	mg.Deps(Vet)

	if err := os.MkdirAll("bin", 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	for name, pkg := range binaries {
		outPath := filepath.Join("bin", name+ext)
		fmt.Printf("==> Compiling %s -> %s\n", pkg, outPath)
		if err := sh.Run("go", "build", "-o", outPath, pkg); err != nil {
			return fmt.Errorf("failed to build %s: %w", name, err)
		}
	}

	fmt.Println("All binaries successfully compiled to bin/")
	return nil
}

// Test runs all unit tests across the workspace
func Test() error {
	fmt.Println("==> Running test suite...")
	return sh.RunV("go", "test", "-v", "./...")
}

// Vet runs the Go static analysis tool on all packages
func Vet() error {
	fmt.Println("==> Running go vet...")
	return sh.Run("go", "vet", "./...")
}

// Fmt formats all Go source files according to Go standards
func Fmt() error {
	fmt.Println("==> Running go fmt...")
	return sh.Run("go", "fmt", "./...")
}

// Clean dynamically parses .gitignore and purges all matching files/directories,
// as well as removing the .git footprint to prepare a clean Moodle submission.
func Clean() error {
	fmt.Println("==> Sanitizing workspace for Moodle submission...")

	patterns, err := parseGitignore(".gitignore")
	if err != nil {
		fmt.Printf("Warning: could not read .gitignore: %v (falling back to default rules)\n", err)
		patterns = []string{"bin/", "logs/", "*.exe", "*.test", "*.out", "*.zip", "ssl-key.log", "*.qlog", "*.sqlog", ".DS_Store", "Thumbs.db", ".vscode/", ".idea/"}
	}

	// Always ensure bin, logs, executables, and zip archives are targeted
	patterns = append(patterns, "bin/", "logs/", "*.exe", "*.zip")

	cleanedCount := 0

	// Walk from current directory
	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if path == "." {
			return nil
		}

		// Never delete critical repo source files
		normalized := filepath.ToSlash(path)
		cleanBase := filepath.Base(path)
		if normalized == ".gitignore" || cleanBase == ".gitignore" ||
			normalized == "go.mod" || cleanBase == "go.mod" ||
			normalized == "go.sum" || cleanBase == "go.sum" ||
			normalized == "magefile.go" || cleanBase == "magefile.go" ||
			normalized == "mage.go" || cleanBase == "mage.go" ||
			normalized == "project-spec.md" || cleanBase == "project-spec.md" {
			return nil
		}

		if matchesAny(normalized, d.Name(), d.IsDir(), patterns) {
			fmt.Printf("Removing: %s\n", normalized)
			if removeErr := os.RemoveAll(path); removeErr == nil {
				cleanedCount++
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Warning during file purge: %v\n", err)
	}

	// Remove .git directory footprint for Moodle upload size optimization
	if _, err := os.Stat(".git"); err == nil {
		fmt.Println("Removing: .git/ repository footprint")
		_ = os.RemoveAll(".git")
		cleanedCount++
	}

	fmt.Printf("Clean complete: %d items removed. Directory is sanitized for Moodle upload.\n", cleanedCount)
	return nil
}

func parseGitignore(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

func matchesAny(relPath, baseName string, isDir bool, patterns []string) bool {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		p = filepath.ToSlash(p)

		// Directory-only pattern (e.g. "bin/", "logs/", ".idea/")
		if strings.HasSuffix(p, "/") {
			dirPattern := strings.TrimSuffix(p, "/")
			if isDir {
				if baseName == dirPattern || relPath == dirPattern || strings.HasPrefix(relPath, dirPattern+"/") {
					return true
				}
			}
			continue
		}

		// Glob pattern (e.g. "*.exe", "*.sqlog", "*~")
		if strings.Contains(p, "*") {
			if matched, _ := filepath.Match(p, baseName); matched {
				return true
			}
			if matched, _ := filepath.Match(p, relPath); matched {
				return true
			}
		}

		// Exact match against base name or relative path
		if baseName == p || relPath == p {
			return true
		}
	}
	return false
}

// Sanity runs compile-time contract checks and verifies that all targets build cleanly
func Sanity() error {
	fmt.Println("==> Step 1: Checking compile-time specification contracts...")
	if err := sh.RunV("go", "test", "-run", "TestContract", "./..."); err != nil {
		return fmt.Errorf("contract verification failed: required variables or functions were modified or missing")
	}

	fmt.Println("\n==> Step 2: Running static analysis and compilation (Build)...")
	if err := Build(); err != nil {
		return err
	}

	fmt.Println("\n[SUCCESS] All specification contracts satisfied and all targets compiled cleanly!")
	return nil
}

// Check is an alias for Sanity
func Check() error {
	return Sanity()
}

// Package verifies the codebase and archives it into H01_<StudentID>.zip
func Package(studentID string) error {
	id := strings.TrimSpace(studentID)
	if id == "" {
		return fmt.Errorf("student ID is required: usage: go run mage.go package <StudentID>")
	}

	if err := Sanity(); err != nil {
		return fmt.Errorf("packaging aborted: codebase does not pass sanity checks")
	}

	// Reuse Clean to sanitize the workspace and purge build/dev artifacts
	if err := Clean(); err != nil {
		return fmt.Errorf("packaging aborted: workspace sanitization failed: %w", err)
	}

	zipName := fmt.Sprintf("H01_%s.zip", id)

	fmt.Printf("\n==> Creating submission package: %s\n", zipName)
	if err := createSubmissionZip(zipName); err != nil {
		return fmt.Errorf("failed to create package %s: %w", zipName, err)
	}

	fmt.Printf("\n[SUCCESS] Package '%s' is ready for submission.\n", zipName)
	return nil
}

func createSubmissionZip(zipName string) error {
	zipFile, err := os.Create(zipName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	const rootDir = "compnet-socket-labs"

	// Create root directory entry
	if _, err := w.Create(rootDir + "/"); err != nil {
		return err
	}

	return filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if path == "." {
			return nil
		}

		normalized := filepath.ToSlash(path)
		// Avoid including zip archives
		if strings.HasSuffix(normalized, ".zip") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = rootDir + "/" + normalized
		if d.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
