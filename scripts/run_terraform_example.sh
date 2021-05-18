#!/bin/bash

if [ -z $ARGO_CLOUDOPS_BUILD_BUCKET ]; then
    echo "Error: ARGO_CLOUDOPS_BUILD_BUCKET envronment variable not set"
    exit 1
fi

if [ -z $ARGO_CLOUDOPS_PROJECT_NAME ]; then
    export ARGO_CLOUDOPS_PROJECT_NAME=project1
fi

if [ -z $ARGO_CLOUDOPS_TARGET_NAME ]; then
    export ARGO_CLOUDOPS_TARGET_NAME=target1
fi

if [ -z $VAULT_ADDR ]; then
    # if not set assume local development
    export VAULT_ADDR="http://docker.for.mac.localhost:8200"
fi

if [ -z $ARGO_CLOUDOPS_SERVICE_ADDR ]; then
    # if not set assume local development
    export ARGO_CLOUDOPS_SERVICE_ADDR=https://localhost:8443
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
CLI=$SCRIPT_DIR/../build/argo-cloudops

code_uri="s3://$ARGO_CLOUDOPS_BUILD_BUCKET/terraform-example.tar.gz"
execute_container_image_uri='argocloudops/argo-cloudops-terraform:0.15.1'
workflow_name=argo-cloudops-single-step-vault-aws

$CLI workflow --type sync \
    -a init='-no-color',execute='-auto-approve -no-color' \
    -e AWS_REGION="us-west-2",CODE_URI="$code_uri",VAULT_ADDR="$VAULT_ADDR" \
    -f terraform \
    -n $ARGO_CLOUDOPS_PROJECT_NAME \
    -p execute_container_image_uri="$execute_container_image_uri" \
    -t $ARGO_CLOUDOPS_TARGET_NAME \
    -w $workflow_name
