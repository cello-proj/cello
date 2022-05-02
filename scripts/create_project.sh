#!/bin/bash -e

GIT_REPO=$1
ACCOUNT_ID=$2

ROLE_NAME=CelloSampleRole

if [ -z $CELLO_PROJECT_NAME ]; then
    export CELLO_PROJECT_NAME=project1
fi

if [ -z $CELLO_TARGET_NAME ]; then
    export CELLO_TARGET_NAME=target1
fi

if [ -z $CELLO_SERVICE_ADDR ]; then
    export CELLO_SERVICE_ADDR=https://localhost:8443
fi

if [ -z $ACCOUNT_ID ]; then
    echo "ACCOUNT_ID for role not set, setting ACCOUNT_ID to current account."
    export ACCOUNT_ID=`aws sts get-caller-identity --query Account --output text`
fi

if [ -z $ACCOUNT_ID ]; then
    echo "Error: Unable to set account id."
    exit 1
fi

if [ -z $GIT_REPO ]; then
    echo "Error: Git repo not set."
    exit 1
fi

set -e

echo "Creating project with target in AWS account '$ACCOUNT_ID' role '$ROLE_NAME'."

cat > /tmp/create_project_request.json<<EOF
{
  "name": "$CELLO_PROJECT_NAME",
  "repository": "$GIT_REPO"
}
EOF

echo "Creating project '$CELLO_PROJECT_NAME'."
output=`curl -s -k \
    -w "\n%{http_code}" \
    -d @/tmp/create_project_request.json \
    -H "Authorization: vault:admin:$CELLO_ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    $CELLO_SERVICE_ADDR/projects`

status_code=`echo "$output" | tail -n1`
if [ $status_code == 400 ]; then 
    echo "ERROR: Project already exists"
    echo "$output" | sed \$d
    exit 1
fi

if [ $status_code != 200 ]; then 
    echo "ERROR: create project failed"
    echo "$output" | sed \$d
    exit 1
fi

response=`echo "$output" | head -1`
token=`jq -n "$response" | jq .token`

cat > /tmp/create_target_request.json<<EOF
{
    "name": "$CELLO_TARGET_NAME",
    "type": "aws_account",
    "properties": {
        "credential_type": "assumed_role",
        "policy_arns": [],
        "role_arn": "arn:aws:iam::$ACCOUNT_ID:role/$ROLE_NAME"
    }
}
EOF

echo "Creating target '$CELLO_TARGET_NAME'."
result=`curl -s -k \
    -o $TMPDIR/response.txt \
    -w "%{http_code}" \
    -d @/tmp/create_target_request.json \
    -H "Authorization: vault:admin:$CELLO_ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    $CELLO_SERVICE_ADDR/projects/$CELLO_PROJECT_NAME/targets`

target_status_code=`echo "$result" | tail -n1`
if [ $target_status_code != 200 ]; then 
    echo "ERROR: create target failed"
    echo "$result" | sed \$d
    exit 1
fi

echo
echo "export CELLO_USER_TOKEN=$token"
echo
