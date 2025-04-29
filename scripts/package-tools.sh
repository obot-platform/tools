#!/bin/bash
set -e -x -o pipefail

REPO=github.com/obot-platform/tools
REPO_DIR=/obot-tools/tools
REPO_NAME=$(basename $REPO_DIR)

if [[ -x "${REPO_DIR}/scripts/build.sh" ]]; then
    (
        echo "Running build script for ${REPO}..."
        cd "${REPO_DIR}"
        ./scripts/build.sh
        echo "Build script for ${REPO} complete!"
    )
else
    echo "No build script found in ${REPO}"
fi

OBOT_SERVER_VERSIONS="$(
    cat <<VERSIONS
${REPO}=$(cd "tools" && git rev-parse --short HEAD),${OBOT_SERVER_VERSIONS}
VERSIONS
)"
OBOT_SERVER_VERSIONS="${OBOT_SERVER_VERSIONS%,}"

cd /
for pj in $(find obot-tools -name package.json | grep -v node_modules); do
    if [ $(basename $(dirname $pj)) == common ]; then
        continue
    fi
    (
        cd $(dirname $pj)
        echo Building $PWD
        pnpm i
    )

done

if ! command -v uv; then
    pip install uv
fi

if [ ! -e obot-tools/venv ]; then
    uv venv /obot-tools/venv
fi

source /obot-tools/venv/bin/activate
uv pip install -r /obot-tools/tools/requirements.txt

cd /obot-tools
cat <<EOF >.envrc.tools.${REPO_NAME}
export GPTSCRIPT_SYSTEM_TOOLS_DIR=/obot-tools/
export GPTSCRIPT_TOOL_REMAP="${REPO}=${REPO_DIR}"
export OBOT_SERVER_TOOL_REGISTRIES="github.com/obot-platform/tools"
export OBOT_SERVER_VERSIONS="${OBOT_SERVER_VERSIONS}"
export TOOLS_VENV_BIN=/obot-tools/venv/bin
EOF
