#/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p ${SCRIPT_DIR}/libs

# Download bats-support
curl -L https://github.com/bats-core/bats-support/archive/master.tar.gz | tar -xzC ${SCRIPT_DIR}/libs
mv ${SCRIPT_DIR}/libs/bats-support-master ${SCRIPT_DIR}/libs/bats-support

# Download bats-assert
curl -L https://github.com/bats-core/bats-assert/archive/master.tar.gz | tar -xzC ${SCRIPT_DIR}/libs
mv ${SCRIPT_DIR}/libs/bats-assert-master ${SCRIPT_DIR}/libs/bats-assert
