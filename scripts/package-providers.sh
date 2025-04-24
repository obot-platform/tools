#!/bin/bash
set -e -x -o pipefail

BIN_DIR=${BIN_DIR:-./bin}

cd /obot-tools

if [ ! -e workspace-provider ]; then
    git clone --depth=1 https://github.com/gptscript-ai/workspace-provider
fi
cd /obot-tools/workspace-provider
go build -ldflags="-s -w" -o bin/gptscript-go-tool .
REGISTRY_REMAP+=('github.com/gptscript-ai/workspace-provider=/obot-tools/workspace-provider')
OBOT_SERVER_VERSIONS="$(
    cat <<VERSIONS
github.com/gptscript-ai/workspace-provider=$(git rev-parse --short HEAD),${OBOT_SERVER_VERSIONS}
VERSIONS
)"

cd /obot-tools

if [ ! -e datasets ]; then
    git clone --depth=1 https://github.com/gptscript-ai/datasets
fi
cd /obot-tools/datasets
go build -ldflags="-s -w" -o bin/gptscript-go-tool .
REGISTRY_REMAP+=('github.com/gptscript-ai/datasets=/obot-tools/datasets')
OBOT_SERVER_VERSIONS="$(
    cat <<VERSIONS
github.com/gptscript-ai/datasets=$(git rev-parse --short HEAD),${OBOT_SERVER_VERSIONS}
VERSIONS
)"

cd /obot-tools

if [ ! -e aws-encryption-provider ]; then
    git clone --depth=1 https://github.com/kubernetes-sigs/aws-encryption-provider
fi
cd /obot-tools/aws-encryption-provider
go build -o "${BIN_DIR}/aws-encryption-provider" cmd/server/main.go
OBOT_SERVER_VERSIONS="$(
    cat <<VERSIONS
github.com/kubernetes-sigs/aws-encryption-provider=$(git rev-parse --short HEAD),${OBOT_SERVER_VERSIONS}
VERSIONS
)"

cd /obot-tools

if [ ! -e kubernetes-kms ]; then
    git clone --depth=1 https://github.com/Azure/kubernetes-kms
fi
cd /obot-tools/kubernetes-kms
go build -ldflags="-s -w" -o "${BIN_DIR}/azure-encryption-provider" cmd/server/main.go
OBOT_SERVER_VERSIONS="$(
    cat <<VERSIONS
github.com/Azure/kubernetes-kms=$(git rev-parse --short HEAD),${OBOT_SERVER_VERSIONS}
VERSIONS
)"
OBOT_SERVER_VERSIONS="${OBOT_SERVER_VERSIONS%,}"

cd /obot-tools

if [ ! -e k8s-cloudkms-plugin ]; then
    git clone --depth=1 https://github.com/obot-platform/k8s-cloudkms-plugin
fi
cd /obot-tools/k8s-cloudkms-plugin
go build -ldflags "-s -w -extldflags 'static'" -installsuffix cgo -tags netgo -o "${BIN_DIR}/gcp-encryption-provider" cmd/k8s-cloudkms-plugin/main.go
OBOT_SERVER_VERSIONS="$(
    cat <<VERSIONS
github.com/obot-platform/k8s-cloudkms-plugin=$(git rev-parse --short HEAD),${OBOT_SERVER_VERSIONS}
VERSIONS
)"

cd /obot-tools
cat <<EOF >.envrc.tools.providers
export GPTSCRIPT_TOOL_REMAP="$(
    IFS=','
    echo "${REGISTRY_REMAP[*]}"
)"
export OBOT_SERVER_VERSIONS="${OBOT_SERVER_VERSIONS}"
EOF
