# Agent Dock

Run Codex or Pi in an isolated Docker container:

```sh
agentdock run codex
agentdock run pi --git https://github.com/example/project.git --branch main
```

## Prerequisites

- macOS or Linux
- Docker Desktop (macOS) or Docker Engine (Linux), running and available as `docker`
- A local harness configuration/auth file: Codex checks `~/.codex/auth.json` and `config.toml`; Pi checks `~/.pi/agent/auth.json`, `models.json`, or `settings.json`
- Go 1.22 to build Agent Dock

The local executable is not used or required. Agent Dock runs the Linux-compatible harness from its harness-specific image and only copies narrowly selected config/auth files into the container.

## Build

```sh
go build -o agentdock ./cmd/agentdock
docker build -t agentdock/codex:0.161.2 -f docker/codex/Dockerfile .
docker build -t agentdock/pi:0.73.1 -f docker/pi/Dockerfile .
```

Images are deliberately not built or pulled implicitly. This makes network and image changes explicit.

## Workspace and sessions

Without `--git`, Agent Dock creates and uses `./workspace`. With `--git`, it clones into a new `workspace/run-*` directory and optionally checks out the requested branch. These directories are retained so work is not lost.

The selected workspace and narrowly selected config/auth files are bind-mounted into the container; the workspace is available at `/workspace`. Codex reads `config.toml` for API-key environment references and mounts a sanitized temporary copy that preserves model/provider routing while removing MCP servers, plugins, and computer-use configuration (including profile overrides). The host configuration is never modified, and the temporary copy is deleted after the run. Codex memory features, memory tools, computer use, bundled app tools, and MCP servers are disabled in the sanitized configuration. Agent Dock does not pass Codex `-c`, `--enable`, or `--disable` overrides, allowing the normal background server mode. The container also masks project `.codex` and `.pi` directories with empty temporary filesystems, preventing repository-local MCP configuration from being loaded. Computer use is disabled for Codex. Agent Dock never mounts a home directory, harness session directory, prior-run volume, Docker socket, or existing session database. Host login sessions are not reused.

API-key environment variables are discovered by reading the selected JSON/TOML configuration files. Pi `auth.json` environment-name keys and `models.json` `$VAR`/`${VAR}` references are resolved from the host environment. Codex TOML/JSON references are handled similarly. Only variables actually referenced by configuration and present on the host are forwarded. Values are written to a short-lived, mode-0600 Docker env file rather than placed in command arguments; literal keys remain in the read-only config mount.

Private Git repositories require credentials available to Git without mounting an entire host home directory. The current MVP does not provide SSH-agent or credential-file forwarding.
