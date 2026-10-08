package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Prepare(root, gitURL, branch string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("workspace root cannot be empty")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	if gitURL == "" {
		return root, nil
	}
	dir, err := os.MkdirTemp(root, "run-")
	if err != nil {
		return "", fmt.Errorf("create repository workspace: %w", err)
	}
	if err := runGit("clone", gitURL, dir); err != nil {
		return "", fmt.Errorf("clone repository: %w", err)
	}
	if branch != "" {
		if err := runGitIn(dir, "checkout", branch); err != nil {
			return "", fmt.Errorf("checkout branch %q: %w", branch, err)
		}
	}
	return dir, nil
}

func runGit(args ...string) error { return runGitIn("", args...) }
func runGitIn(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, output)
	}
	return nil
}
