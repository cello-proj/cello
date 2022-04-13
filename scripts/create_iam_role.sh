#!/bin/bash

set -e 

role_name=CelloSampleRole

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

aws cloudformation create-stack \
    --template-body file://$SCRIPT_DIR/iam_role.json \
    --stack-name $role_name \
    --parameters ParameterKey=RoleName,ParameterValue=$role_name \
    --capabilities CAPABILITY_NAMED_IAM \
    --region us-west-2
