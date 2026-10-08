package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	configpkg "github.com/agentdock/agentdock/internal/config"
	dockerpkg "github.com/agentdock/agentdock/internal/docker"
	"github.com/agentdock/agentdock/internal/harness"
	"github.com/agentdock/agentdock/internal/workspace"
)

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }
func ExitCode(err error) int {
	var e *exitError
	if errors.As(err, &e) {
		return e.code
	}
	return 1
}

func Run(args []string, in io.Reader, out, errOut io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		usage(out)
		return nil
	}
	if args[0] != "run" {
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText)
	}
	return run(args[1:], in, out, errOut)
}

func run(args []string, in io.Reader, out, errOut io.Writer) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(errOut)
	gitURL := fs.String("git", "", "Git repository URL to clone")
	branch := fs.String("branch", "", "branch to check out after cloning")
	workspaceRoot := fs.String("workspace", "", "workspace root (default: ./workspace)")
	imageOverride := fs.String("image", "", "override the selected harness image")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("run requires exactly one harness: codex or pi")
	}
	name := strings.ToLower(fs.Arg(0))
	h, ok := harness.Lookup(name)
	if !ok {
		return fmt.Errorf("unsupported harness %q (supported: codex, pi)", name)
	}
	if *branch != "" && *gitURL == "" {
		return errors.New("--branch requires --git")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}
	localConfigFiles := h.LocalConfigFiles(home)
	configFiles := h.ConfigFiles(home)
	if len(localConfigFiles) == 0 {
		return fmt.Errorf("no %s configuration found; authenticate locally first (session files are never mounted)", h.Name)
	}
	if err := dockerpkg.CheckAvailable(); err != nil {
		return err
	}

	root := *workspaceRoot
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = filepath.Join(cwd, "workspace")
	}
	ws, err := workspace.Prepare(root, *gitURL, *branch)
	if err != nil {
		return err
	}
	image := h.Image
	if *imageOverride != "" {
		image = *imageOverride
	}
	if err := dockerpkg.EnsureImage(image, h.Dockerfile); err != nil {
		return err
	}
	fromConfig, err := configpkg.Environment(h.Name, localConfigFiles)
	if err != nil {
		return err
	}
	if h.Name == "codex" {
		path, cleanup, err := configpkg.PrepareCodex(localConfigFiles)
		if err != nil {
			return err
		}
		defer cleanup()
		if path != "" {
			configFiles = append(configFiles, path)
		}
	}
	return dockerpkg.Run(dockerpkg.Options{Image: image, Workspace: ws, ConfigFiles: configFiles, Environment: fromConfig, In: in, Out: out, ErrOut: errOut, Command: h.ContainerCommand})
}

func usage(w io.Writer) { fmt.Fprint(w, usageText) }

const usageText = `Usage:
  agentdock run <codex|pi> [--git <url>] [--branch <branch>]

Options:
  --git URL          Clone a repository into a new workspace
  --branch NAME      Check out a branch after cloning (requires --git)
  --workspace PATH   Use a different workspace root
  --image IMAGE      Override the harness image

The container receives only the selected workspace and selected config/auth files. Existing harness sessions,
host home directories, and unrelated environment variables are never mounted.
`
