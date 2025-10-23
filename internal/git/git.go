package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// IsGitRepository checks if path is inside a git work tree (supports worktrees)
func IsGitRepository(path string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		// Check if it's just "not a git repository" vs a real error
		stderr := out.String()
		if strings.Contains(stderr, "not a git repository") ||
			strings.Contains(stderr, "Not a git repository") ||
			strings.Contains(stderr, "fatal: not a git repository") {
			return false, nil // This is expected for non-git directories
		}
		// This is a real error (permissions, path doesn't exist, etc.)
		return false, fmt.Errorf("git check failed: %w", err)
	}
	return strings.TrimSpace(out.String()) == "true", nil
}

func GetStatus(path string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	return out.String(), nil
}

// GetDiff returns git diff output for staged or unstaged changes
func GetDiff(path string, staged bool) (string, error) {
	var args []string
	if staged {
		args = []string{"diff", "--staged"}
	} else {
		args = []string{"diff"}
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		// Don't fail on no changes or no HEAD
		if !isNoHeadError(err) {
			return "", fmt.Errorf("git diff: %w", err)
		}
	}
	return out.String(), nil
}

// GetCombinedDiff returns unstaged and staged diffs with headings; tolerant of no HEAD
func GetCombinedDiff(path string) (string, error) {
	var b strings.Builder

	unstaged, err1 := runGit(path, []string{"diff"})
	if err1 != nil && !isNoHeadError(err1) {
		// still include whatever output we have
	}
	staged, _ := runGit(path, []string{"diff", "--staged"})

	if strings.TrimSpace(unstaged) != "" {
		b.WriteString("## Unstaged changes (git diff)\n")
		b.WriteString(unstaged)
		b.WriteString("\n")
	}
	if strings.TrimSpace(staged) != "" {
		b.WriteString("## Staged changes (git diff --staged)\n")
		b.WriteString(staged)
		b.WriteString("\n")
	}
	return b.String(), nil
}

// GetDiffNumStat returns file statistics for changes (additions, deletions, filename)
func GetDiffNumStat(path string) (string, error) {
	cmd := exec.Command("git", "diff", "--numstat", "HEAD")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if !isNoHeadError(err) {
			return "", fmt.Errorf("git diff --numstat: %w", err)
		}
	}
	return out.String(), nil
}

// GetStagedDiffNumStat returns file statistics for staged changes
func GetStagedDiffNumStat(path string) (string, error) {
	cmd := exec.Command("git", "diff", "--numstat", "--staged")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if !isNoHeadError(err) {
			return "", fmt.Errorf("git diff --numstat --staged: %w", err)
		}
	}
	return out.String(), nil
}

func runGit(path string, args []string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

func isNoHeadError(err error) bool {
	return err != nil && (errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "unknown revision or path not in the working tree") || strings.Contains(err.Error(), "ambiguous argument 'HEAD'"))
}

// StageFile stages a specific file
func StageFile(path, filename string) error {
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add %s: %w", filename, err)
	}
	return nil
}

// StageAll stages all untracked and modified files (respects .gitignore)
func StageAll(path string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add -A: %w", err)
	}
	return nil
}

// Commit creates a git commit with the given message
func Commit(path, message string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git commit: %w", err)
	}

	// Get commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = path
	var hashOut bytes.Buffer
	hashCmd.Stdout = &hashOut
	hashCmd.Stderr = &hashOut
	if err := hashCmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(hashOut.String()), nil
}
