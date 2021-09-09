#!/bin/bash

ARGO_CLOUDOPS_VERSION=0.8.1
ARGO_WORKFLOWS_VERSION=3.0.10

set -e

# check if executable exists, if not give command for installing.
# example:
#  executable_check "application" "application install command"

function executable_check()  {
  if ! command -v $1 &> /dev/null
  then
      echo "executable $1 could not be found"
      echo "to install $1 run:"
      echo "  $2"
      exit
  fi
}

executable_check "brew" '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"'
executable_check "docker" 'brew cask install docker'
executable_check "kubectl" 'curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"'
executable_check "argo" 'brew install argo'
executable_check "aws" 'brew install awscli'

# exit if aws credentials are not set
set +e
aws sts get-caller-identity &> /dev/null
if [ $? != 0 ]; then
  echo "aws credentials not set"
  exit 1
fi
set -e

if [ -z $ARGO_CLOUDOPS_ADMIN_SECRET ]; then
  echo "ARGO_CLOUDOPS_ADMIN_SECRET environment variable must be set"
  exit 1
fi

if [ ! -d scripts ]; then
  echo "this script must be run from the root of the repo"
  exit 1
fi

# create iam role if it doesn't already exist
set +e
aws cloudformation describe-stacks --stack-name ArgoCloudOpsSampleRole --region us-west-2 &> /dev/null
if [ $? != 0 ]; then
  echo "creating iam role"
  bash scripts/create_iam_role.sh
fi
set -e

set +e
docker build --pull --rm -f "Dockerfile" -t argocloudops:latest "."
kubectl apply -f ./scripts/quickstart_manifest.yaml
# Sleeping after applying manifest so pods have time to start
sleep 60
set -e

# setup postgres db
# dont fail if alredy exists
set +e
export POSTGRES_POD="$(kubectl get pods --no-headers -o custom-columns=":metadata.name" | grep postgres)"
kubectl exec $POSTGRES_POD -- psql -lqt | cut -d \| -f 1 | grep argocloudops
if [ $? != 0 ]; then
  kubectl cp ./scripts/createdbtables.sql $POSTGRES_POD:./createdbtables.sql
  kubectl exec $POSTGRES_POD -- createdb argocloudops -U postgres
  kubectl exec $POSTGRES_POD -- psql -U postgres -d argocloudops -f ./createdbtables.sql
fi
set -e

# setup workflow if it doesn't exist
set -e
argo template get -n argo argo-cloudops-single-step-vault-aws &> /dev/null
if [ $? != 0 ]; then
  argo template create -n argo workflows/argo-cloudops-single-step-vault-aws.yaml
fi

# setup aws credentials in vault
export AWS_ACCESS_KEY_ID=`aws configure get $AWS_PROFILE.aws_access_key_id`
export AWS_SECRET_ACCESS_KEY=`aws configure get $AWS_PROFILE.aws_secret_access_key`
export AWS_SESSION_TOKEN=`aws configure get $AWS_PROFILE.aws_session_token`

cat > /tmp/awsConfig << EOF
[default]
aws_access_key_id=$AWS_ACCESS_KEY_ID
aws_secret_access_key=$AWS_SECRET_ACCESS_KEY
aws_session_token=$AWS_SESSION_TOKEN
EOF
kubectl exec vault-0 -- mkdir -p /home/vault/.aws
kubectl cp /tmp/awsConfig vault-0:/home/vault/.aws/credentials

echo "Argo Cloudops started, forwarding to port 8443"
export ACO_POD="$(kubectl get pods --no-headers -o custom-columns=":metadata.name" | grep argocloudops)"
kubectl port-forward $ACO_POD 8443:8443


