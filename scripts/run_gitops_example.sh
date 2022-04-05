#!/bin/bash -e

GIT_PATH=$1
GIT_SHA=$2
MODE=$3

if [ -z $ARGO_CLOUDOPS_PROJECT_NAME ]; then
    export ARGO_CLOUDOPS_PROJECT_NAME=project1
fi

if [ -z $ARGO_CLOUDOPS_TARGET_NAME ]; then
    export ARGO_CLOUDOPS_TARGET_NAME=target1
fi

if [ -z $ARGO_CLOUDOPS_SERVICE_ADDR ]; then
    # if not set assume local development
    export ARGO_CLOUDOPS_SERVICE_ADDR=https://localhost:8443
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
if [ "${MODE}" == "dev" ]; then
  CLI=$SCRIPT_DIR/../build/argo-cloudops
else
  CLI=$SCRIPT_DIR/../quickstart/argo-cloudops
fi

$CLI exec \
    -n $ARGO_CLOUDOPS_PROJECT_NAME \
    -t $ARGO_CLOUDOPS_TARGET_NAME \
    -p $GIT_PATH \
    -s $GIT_SHA
