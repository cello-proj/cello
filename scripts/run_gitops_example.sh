#!/bin/bash -e

GIT_PATH=$1
GIT_SHA=$2
MODE=$3

if [ -z $CELLO_PROJECT_NAME ]; then
    export CELLO_PROJECT_NAME=project1
fi

if [ -z $CELLO_TARGET_NAME ]; then
    export CELLO_TARGET_NAME=target1
fi

if [ -z $CELLO_SERVICE_ADDR ]; then
    # if not set assume local development
    export CELLO_SERVICE_ADDR=https://localhost:8443
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
if [ "${MODE}" == "dev" ]; then
  CLI=$SCRIPT_DIR/../build/cello
else
  CLI=$SCRIPT_DIR/../quickstart/cello
fi

$CLI exec \
    -n $CELLO_PROJECT_NAME \
    -t $CELLO_TARGET_NAME \
    -p $GIT_PATH \
    -s $GIT_SHA
