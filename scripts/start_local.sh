#!/bin/bash

unset VAULT_TOKEN
export VAULT_ADDR='http://127.0.0.1:8200'
export ARGO_ADDR='http://127.0.0.1:9000'

MODE=${1}

if [ -z "$CELLO_DB_HOST" ]; then
    export CELLO_DB_HOST=localhost
fi

if [ -z "$CELLO_DB_NAME" ]; then
    export CELLO_DB_NAME=cello
fi

if [ -z "$CELLO_DB_USER" ]; then
    export CELLO_DB_USER=cello
fi

if [ -z "$CELLO_DB_PASSWORD" ]; then
    export CELLO_DB_PASSWORD=1234
fi

if [ -z "$CELLO_GIT_AUTH_METHOD" ]; then
    export CELLO_GIT_AUTH_METHOD=https
fi

# Vault was not loading credentials from the default chain, try to fetch from profile
if [ -n "${AWS_PROFILE}" ]; then
    CREDS_PROCESS_VALUE=`aws configure get $AWS_PROFILE.credential_process`
    if [ -n "$CREDS_PROCESS_VALUE" ]; then
        # profile is using credential_process 
        # we require jq make sure it exits
        if ! command -v jq >/dev/null 2>&1; then
            echo "ERROR: 'jq' command could not be found"
            exit 1
        fi

        # run credential_process - parse json from it
        CREDS_JSON=`$CREDS_PROCESS_VALUE`

        export AWS_ACCESS_KEY_ID=`echo "$CREDS_JSON" | jq -r ."AccessKeyId"`
        export AWS_SECRET_ACCESS_KEY=`echo "$CREDS_JSON" | jq -r ."SecretAccessKey"`
        export AWS_SESSION_TOKEN=`echo "$CREDS_JSON" | jq -r ."SessionToken"`
    else
        # profile isn't using credential_process, get values from profile config
        export AWS_ACCESS_KEY_ID=`aws configure get $AWS_PROFILE.aws_access_key_id`
        export AWS_SECRET_ACCESS_KEY=`aws configure get $AWS_PROFILE.aws_secret_access_key`
        export AWS_SESSION_TOKEN=`aws configure get $AWS_PROFILE.aws_session_token`
    fi
fi

if [ -z "$CELLO_CONFIG" ]; then
    export CELLO_CONFIG=cello.yaml
fi

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "Error: AWS_ACCESS_KEY_ID was not set and could not be loaded from AWS_PROFILE"
    exit 1
fi

if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "Error: AWS_SECRET_ACCESS_KEY was not set and could not be loaded from AWS_PROFILE"
    exit 1
fi

if [ -z "$AWS_SESSION_TOKEN" ]; then
    echo "Error: AWS_SESSION_TOKEN was not set and could not be loaded from AWS_PROFILE"
    exit 1
fi

set -e

# get account ID
export ACCOUNT_ID=`aws sts get-caller-identity --query Account --output text`
if [ -z "$ACCOUNT_ID" ]; then
    echo "Account ID not found!"
    exit 1
fi

echo "Starting with credentials in AWS account '$ACCOUNT_ID'."

pkill -9 vault && true

sleep 2

vault server -dev &

sleep 2

echo "Vault starting, hang in there..."

sleep 5

export VAULT_TOKEN=`cat ~/.vault-token`

vault secrets enable aws
vault auth enable approle

cat > /tmp/argo-cloudops-policy.hcl << EOF
# Create and manage roles
path "auth/approle/role/argo-cloudops-projects-*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}

# Write ACL policies
path "sys/policies/acl/argo-cloudops-projects-*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}

# Write AWS roles
path "aws/roles/argo-cloudops-projects-*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}

# List roles
path "aws/roles/*" {
  capabilities = [ "read", "list" ]
}
EOF

vault policy write argo-cloudops-service /tmp/argo-cloudops-policy.hcl

vault write auth/approle/role/argo-cloudops policies=argo-cloudops-service secret_id_ttl=8766h

CELLO_VAULT_ROLE_ID=$(vault read -format json \
    auth/approle/role/argo-cloudops/role-id \
    | jq -r '.data.role_id')

CELLO_VAULT_SECRET_ID=$(vault write -f -format json \
    auth/approle/role/argo-cloudops/secret-id \
    | jq -r '.data.secret_id')

export service_vault_token=$(vault write --format json \
    auth/approle/login \
    role_id="${CELLO_VAULT_ROLE_ID}" \
    secret_id="${CELLO_VAULT_SECRET_ID}" \
    | jq -r '.auth.client_token')

# Set to env values expected by service to start
export VAULT_ROLE=${CELLO_VAULT_ROLE_ID}
export VAULT_SECRET=${CELLO_VAULT_SECRET_ID}
export VAULT_TOKEN=$service_vault_token

mkdir -p ./ssl
# generate certificate, suppress output unless there is an error
output=$(openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 -subj "/C=US/ST=CA/L=Mountain View/O=Cognition/CN=app" -keyout ./ssl/certificate.key -out ./ssl/certificate.crt 2>&1) || echo "$output"

echo "Starting Cello Service"
if [ "${MODE}" == "dev" ]; then
  build/service
else
  quickstart/service
fi
echo "Cello Serivce Stopped"
