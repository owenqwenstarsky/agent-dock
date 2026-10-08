# Isolation and sessions

Agent Dock treats each run as a fresh harness session. The Docker invocation bind-mounts the prepared workspace and only the selected Codex/Pi auth/config files. It does not mount `$HOME` wholesale, any session database, or a prior container volume. It also does not forward the host environment wholesale.

Existing host sessions are intentionally unavailable inside the container. Selected auth files are read-only mounts; session files are never selected. Do not work around isolation by mounting a home directory.

The container runs as a non-root user. The workspace is retained on the host, including Git clones created under `workspace/run-*`.
