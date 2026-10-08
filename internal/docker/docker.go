package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

type Options struct {
	Image, Workspace string
	ConfigFiles      []string
	Environment      map[string]string
	Command          []string
	In               io.Reader
	Out, ErrOut      io.Writer
}

func CheckAvailable() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("Docker CLI is not installed; install Docker Desktop or Docker Engine")
	}
	cmd := exec.Command("docker", "info")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Docker daemon is unavailable: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func EnsureImage(image, dockerfile string) error {
	inspect := exec.Command("docker", "image", "inspect", image)
	if inspect.Run() == nil {
		return nil
	}
	return fmt.Errorf("image %s is not available; build it with: docker build -t %s -f %s .", image, image, dockerfile)
}

func Run(o Options) error {
	envFile, err := environmentFile(o.Environment)
	if err != nil {
		return err
	}
	if envFile != "" {
		defer os.Remove(envFile)
	}
	args := []string{"run", "--rm", "-it", "--init", "--mount", "type=bind,src=" + o.Workspace + ",dst=/workspace", "--mount", "type=tmpfs,dst=/workspace/.codex", "--mount", "type=tmpfs,dst=/workspace/.pi"}
	hasCodexConfig := false
	for _, source := range o.ConfigFiles {
		destination := configDestination(source)
		if destination == "/home/agent/.codex/config.toml" {
			// Mount only a seed: Codex persists trust using atomic replacement,
			// which cannot work on a read-only (or single-file writable) mount.
			destination = "/tmp/agentdock-codex-config.toml"
			hasCodexConfig = true
		}
		args = append(args, "--mount", "type=bind,readonly,src="+source+",dst="+destination)
	}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	if hasCodexConfig {
		args = append(args, "--entrypoint", "/bin/sh")
	}
	args = append(args, "--workdir", "/workspace", o.Image)
	if hasCodexConfig {
		args = append(args, "-c", codexBootstrap, "agentdock")
	}
	args = append(args, o.Command...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = o.In, o.Out, o.ErrOut
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		for sig := range signals {
			_ = cmd.Process.Signal(sig)
		}
	}()
	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: e.ExitCode(), Err: fmt.Errorf("container exited with status %d", e.ExitCode())}
		}
		return err
	}
	return nil
}

// Config remains container-local and writable for trust decisions. Never run
// interpolated config values as shell code; harness arguments use "$@".
const codexBootstrap = `umask 077; mkdir -p /home/agent/.codex && cp /tmp/agentdock-codex-config.toml /home/agent/.codex/config.toml && chmod 600 /home/agent/.codex/config.toml && exec codex "$@"`

func environmentFile(environment map[string]string) (string, error) {
	if len(environment) == 0 {
		return "", nil
	}
	keys := make([]string, 0, len(environment))
	for key := range environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	file, err := os.CreateTemp("", "agentdock-env-*.env")
	if err != nil {
		return "", fmt.Errorf("create temporary environment file: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", err
	}
	for _, key := range keys {
		// Credentials loaded from command-line tools sometimes retain a trailing
		// line ending. Docker env files cannot represent embedded newlines, but
		// trailing ones are not part of the credential and are safe to remove.
		value := strings.TrimRight(environment[key], "\r\n")
		if strings.ContainsAny(value, "\r\n") || strings.ContainsAny(key, "\r\n=") {
			file.Close()
			os.Remove(file.Name())
			return "", fmt.Errorf("API key environment value for %s contains an unsupported embedded newline", key)
		}
		if _, err := fmt.Fprintf(file, "%s=%s\n", key, value); err != nil {
			file.Close()
			os.Remove(file.Name())
			return "", err
		}
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

func configDestination(source string) string {
	if i := strings.Index(source, "/.codex/"); i >= 0 {
		return "/home/agent/.codex/" + source[i+len("/.codex/"):]
	}
	if i := strings.Index(source, "/.pi/agent/"); i >= 0 {
		return "/home/agent/.pi/agent/" + source[i+len("/.pi/agent/"):]
	}
	return "/home/agent/config"
}

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
