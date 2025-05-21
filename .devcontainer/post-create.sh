#!/bin/bash

set -eou pipefail

sudo sed -i 's/^#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config
npm install -g @anthropic-ai/claude-code

python -m pip install aider-install aider-chat
aider-install
