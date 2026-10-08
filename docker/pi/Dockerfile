FROM node:22.14.0-bookworm-slim
ARG PI_VERSION=0.73.1
RUN apt-get update \
    && apt-get install --no-install-recommends --yes git procps ripgrep \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global @mariozechner/pi-coding-agent@${PI_VERSION} \
    && useradd --create-home --uid 10001 agent \
    && mkdir -p /home/agent/.pi/agent \
    && chown -R agent:agent /home/agent
USER agent
WORKDIR /workspace
ENTRYPOINT ["pi"]
