FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# ensure ssh dir exists to permissions aren't wonky when we mount authorized keys
RUN mkdir -p /home/vscode/.ssh && chown vscode:vscode /home/vscode/.ssh
