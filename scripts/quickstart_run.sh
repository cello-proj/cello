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
executable_check "kubectl" 'curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"'
executable_check "argo" 'brew install argo'
executable_check "aws" 'brew install awscli'
executable_check "psql" "brew install postgresql"
executable_check "createdb" "brew install postgresql"
executable_check "curl" 'brew install curl'

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

# create argo namespace if it doesn't exist
set +e
kubectl get namespace argo &> /dev/null
if [ $? != 0 ]; then
  echo "creating argo namespace"
  kubectl create namespace argo
fi
set -e

# install argo workflows from manifext
echo "installing/updating argo"
kubectl apply -n argo -f https://raw.githubusercontent.com/argoproj/argo-workflows/v${ARGO_WORKFLOWS_VERSION}/manifests/install.yaml &> /dev/null


# setup postgres db
# dont fail if alredy exists
set +e
createdb argocloudops &> /dev/null
set -e
psql -d argocloudops -f scripts/createdbtables.sql &> /dev/null


# setup workflow if it doesn't exist
set -e
argo template get -n argo argo-cloudops-single-step-vault-aws &> /dev/null
if [ $? != 0 ]; then
  argo template create -n argo workflows/argo-cloudops-single-step-vault-aws.yaml
fi

mkdir -p quickstart

# download Argo Cloudops CLI if it doesn't exist
if [ ! -f quickstart/argo-cloudops ]; then
    curl -L https://github.com/argoproj-labs/argo-cloudops/releases/download/v${ARGO_CLOUDOPS_VERSION}/argo-cloudops_cli_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz -o quickstart/argo-cloudops_cli_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz &> /dev/null
      tar -xzf quickstart/argo-cloudops_cli_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz -C quickstart/ &> /dev/null
        rm quickstart/argo-cloudops_cli_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz &> /dev/null
fi

# download Argo CloudOps service binary if it doesn't exist
if [ ! -f quickstart/service ]; then
  curl -L https://github.com/argoproj-labs/argo-cloudops/releases/download/v${ARGO_CLOUDOPS_VERSION}/argo-cloudops_service_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz -o quickstart/argo-cloudops_service_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz &> /dev/null
  tar -xzf quickstart/argo-cloudops_service_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz -C quickstart/ &> /dev/null
  rm quickstart/argo-cloudops_service_${ARGO_CLOUDOPS_VERSION}_darwin_x86_64.tar.gz &> /dev/null
fi

echo "Starting Argo Cloudops service, use Ctrl+c to end process"
./scripts/start_local.sh


