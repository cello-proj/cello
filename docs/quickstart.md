# Quickstart

Note: this is a quick guide for getting something up and running. This is configured for local setups and not meant to be run in production

## Pre-reqs

_The quickstart currently only supports macOS._

- Clone the Cello GitHub [repository](https://github.com/cello-proj/cello).

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop), ensure [Kubernetes](https://docs.docker.com/desktop/kubernetes/) is running.

- Install [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)

- Install **Argo CLI** `brew install argo`

- Install [jq](https://stedolan.github.io/jq/) for json parsing.

## Deploy Sample App Locally

You will need two windows

1. Vault & Cello Service
1. Client commands, etc

### Start Vault & Cello Service

- In window **#1**, ensure you have [AWS credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html) for the target account configured and access to your kubernetes cluster. For the AWS credentials, export the AWS_PROFILE that is to be used.

- Set **CELLO_ADMIN_SECRET** env var to `abcd1234abcd1234`.

  ```sh
  export CELLO_ADMIN_SECRET=abcd1234abcd1234
  ```

- Start the Cello Service (includes workflows, vault, and postgres).
  Note: this will copy your current AWS credentials to the vault containers.

  ```sh
  bash scripts/quickstart_run.sh
  ```

### Create Cello Project And Target (One Time Setup)

- In window **#2**, ensure you have the **CELLO_ADMIN_SECRET**
  env var set to `abcd1234abcd1234`.

      ```sh
      export CELLO_ADMIN_SECRET=abcd1234abcd1234
      ```

- Ensure your AWS credentials are set for the **target account** and create
  your first project and target. The output contains an export command for the **CELLO_USER_TOKEN**
  for the new project.

  ```sh
  bash scripts/create_project.sh https://github.com/cello-proj/cello.git
  ```

### Run Workflow

- In window **#2**, ensure the **CELLO_USER_TOKEN** for the project is
  specified (the output of `create_project.sh` should have output a bash
  command to export it).

- CDK Example

  ```sh
  # CDK Example
  CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/kube_cdk_manifest.yaml e3a419e69a5ae762862dc7cf382304a4e6cc2547`

  # Get the status/follow the logs
  ./quickstart/cello get $CDK_WORKFLOW_NAME
  ./quickstart/cello logs -f $CDK_WORKFLOW_NAME
  ```

- TERRAFORM Example

  ```sh
  # Terraform Example
  TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/kube_terraform_manifest.yaml e3a419e69a5ae762862dc7cf382304a4e6cc2547`

  # Get the status/follow the logs
  ./quickstart/cello get $TERRAFORM_WORKFLOW_NAME
  ./quickstart/cello logs -f $TERRAFORM_WORKFLOW_NAME
  ```
