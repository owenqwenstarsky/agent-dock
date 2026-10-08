FROM node:22.14.0-bookworm-slim
ARG CODEX_VERSION=0.161.0
RUN apt-get update \
    && apt-get install --no-install-recommends --yes bubblewrap git procps ripgrep \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global @openai/codex@${CODEX_VERSION} \
    && useradd --create-home --uid 10001 agent \
    && mkdir -p /home/agent/.codex \
    && chown -R agent:agent /home/agent
USER agent
WORKDIR /workspace
ENTRYPOINT ["codex"]
