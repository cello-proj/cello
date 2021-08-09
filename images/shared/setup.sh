#!/bin/bash

# This is a sample script which exchanges a vault token, downloads zip
# archive from S3 for a given project and target.
#
# This script assumes that the zip file is available via the same account
# credentials which are used to run the framework.

credentials_file=/root/.aws/credentials

export VAULT_TOKEN=$1
export PROJECT_NAME=$2
export TARGET_NAME=$3

usage() {
    echo
    echo "$0 VAULT_TOKEN PROJECT_NAME TARGET_NAME"
    echo
    echo "CODE_URI env variable must be set with S3 uri for zip archive "
    echo "VAULT_ADDR env variable must have valid vault endpoint"
    echo
}

if [ -z $CODE_URI ]; then
    echo "Error: CODE_URI not supplied"
    usage
    exit 1
fi

if [ -z $PROJECT_NAME ]; then
    echo "Error: PROJECT_NAME not set"
    usage
    exit 1
fi

if [ -z $TARGET_NAME ]; then
    echo "Error: TARGET_NAME not set"
    usage
    exit 1
fi

if [ -z $VAULT_ADDR ]; then
    echo "Error: VAULT_ADDR not set"
    usage
    exit 1
fi

if [ -z $VAULT_TOKEN ]; then
    echo "Error: VAULT_TOKEN not set"
    usage
    exit 1
fi

set -euo pipefail

echo "Starting setup."
echo "CODE_URI: $CODE_URI"
echo "PROJECT_NAME: $PROJECT_NAME"
echo "TARGET_NAME: $TARGET_NAME"
echo "VAULT_ADDR: $VAULT_ADDR"

#
# Get credentials from vault
#
vault_project_prefix='aws/sts/argo-cloudops'
target="${vault_project_prefix}-projects-${PROJECT_NAME}-target-${TARGET_NAME}"

token_head=`echo $VAULT_TOKEN |cut -b1-8`
echo "Exchanging token '${token_head}...' via '$VAULT_ADDR' for target '$target'"

creds=$(vault read --format json $target | \
    jq -r '"aws_access_key_id=\(.data.access_key)\naws_secret_access_key=\(.data.secret_key)\naws_session_token=\(.data.security_token)"')

echo "Exchanging token successful."

echo "Writing credentials to '$credentials_file'."
cat > $credentials_file <<EOF
[default]
$creds
EOF

arn=`aws sts get-caller-identity --output text --query Arn`
echo "Arn of role assumed '$arn'."

if [[ "$CODE_URI" =~ ^s3://.* ]]; then
    echo "Downloading $CODE_URI from S3."
    aws s3 cp $CODE_URI code.tar.gz
    echo "Download complete."
elif [[ "$CODE_URI" =~ ^https://.* ]]; then
    echo "Downloading $CODE_URI using HTTPS."
    curl -L -o code.tar.gz ${CODE_URI}
    echo "Download complete."
else
    echo "$CODE_URI is not a supported format"
    exit 1
fi

echo "Decompressing"
tar -xzf code.tar.gz
echo "Decompressing complete."

echo "Setup complete."
