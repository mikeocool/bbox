#!/bin/bash

set -eou pipefail

# fix ssh dir permissions
sudo chown ${USER}:${USER} ~/.ssh

sudo sed -i 's/^#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config
npm install -g @anthropic-ai/claude-code
