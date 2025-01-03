FROM cgr.dev/chainguard/wolfi-base
WORKDIR /mnt/data
RUN apk add -U kubectl helm jq wget curl procps openssh-client bash coreutils k3d
RUN apk add -U git gh
COPY --chmod=0777 <<"EOF" /run.sh
#!/bin/bash
set -x
if [ -n "$KUBECONFIG" ]; then
    mkdir $HOME/.kube
    echo "$KUBECONFIG" | sed 's/[\n ]//g' | base64 -d > $HOME/.kube/config
    export KUBECONFIG=$HOME/.kube/config
fi
bash -c "${CMD}"
EOF
CMD ["/run.sh"]
