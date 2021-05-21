#!/bin/bash

unset VAULT_TOKEN
export VAULT_ADDR='http://127.0.0.1:8200'
export ARGO_ADDR='http://127.0.0.1:9000'

if [ -z "$DB_HOST" ]; then
    export DB_HOST=localhost
fi

if [ -z "$DB_DATABASE_NAME" ]; then
    export DB_DATABASE_NAME=argocloudops
fi

if [ -z "$DB_USER" ]; then
    export DB_USER=argoco
fi

if [ -z "$DB_PASSWORD" ]; then
    export DB_PASSWORD=1234
fi

if [ -z "$SSH_PEM_FILE" ]; then
    export SSH_PEM_FILE=$HOME/.ssh/id_rsa
fi

# TODO / HACK: Vault was not loading credentials from the default chain.
if [ -n "${AWS_PROFILE}" ]; then
    export AWS_ACCESS_KEY_ID=`aws configure get $AWS_PROFILE.aws_access_key_id`
    export AWS_SECRET_ACCESS_KEY=`aws configure get $AWS_PROFILE.aws_secret_access_key`
    export AWS_SESSION_TOKEN=`aws configure get $AWS_PROFILE.aws_session_token`
fi

if [ -z "$ARGO_CLOUDOPS_CONFIG" ]; then
    export ARGO_CLOUDOPS_CONFIG=argo-cloudops.yaml
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

export ACCOUNT_ID=`aws sts get-caller-identity --query Account --output text`
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

ARGO_CLOUDOPS_VAULT_ROLE_ID=$(vault read -format json \
    auth/approle/role/argo-cloudops/role-id \
    | jq -r '.data.role_id')

ARGO_CLOUDOPS_VAULT_SECRET_ID=$(vault write -f -format json \
    auth/approle/role/argo-cloudops/secret-id \
    | jq -r '.data.secret_id')

export service_vault_token=$(vault write --format json \
    auth/approle/login \
    role_id="${ARGO_CLOUDOPS_VAULT_ROLE_ID}" \
    secret_id="${ARGO_CLOUDOPS_VAULT_SECRET_ID}" \
    | jq -r '.auth.client_token')

# Set to env values expected by service to start
export VAULT_ROLE=${ARGO_CLOUDOPS_VAULT_ROLE_ID}
export VAULT_SECRET=${ARGO_CLOUDOPS_VAULT_SECRET_ID}
export VAULT_TOKEN=$service_vault_token

mkdir -p ./ssl
# generate certificate, suppress output unless there is an error
output=$(openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 -subj "/C=US/ST=CA/L=Mountain View/O=Cognition/CN=app" -keyout ./ssl/certificate.key -out ./ssl/certificate.crt 2>&1) || echo "$output"

echo "Starting Argo CloudOps Service"
build/service
echo "Argo CloudOps Serivce Stopped"
