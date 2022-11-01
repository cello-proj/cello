# Development Environment setup

## Pre-reqs

The below instructions assume Cello is on your local OSX system
with Docker Desktop managing resource in AWS (region us-west-2) with credentials provided by Vault.

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop), ensure kubernetes is running.

- Install [Argo Workflows](https://argoproj.github.io/argo-workflows/installation/)

- Install **Argo CLI** `brew install argo`

- Install **GoLang** `brew install golang`

- Install GoLint `go get -u golang.org/x/lint/golint` and ensure `$GOPATH` is in your `$PATH`.

- Install PostgreSQL `brew install postgresql`

- Install [golang-migrate](https://github.com/golang-migrate/migrate)

- Install [Vault](https://www.vaultproject.io/downloads) for credential generation.

- Install [jq](https://stedolan.github.io/jq/) for json parsing.

- Install **npm** `brew install npm` (For CDK).

- Install [terraform](https://www.terraform.io/downloads.html).

## Validate argo workflows is setup and working correctly.

- Submit Argo Hello World workflow and record the **Name** from the output.

  ```
  argo submit -n argo https://raw.githubusercontent.com/argoproj/argo-workflows/master/examples/hello-world.yaml
  ```

- Ensure the workflow completes with Status **Succeeded**.

  ```
  argo get -n argo <UPDATE_WITH_NAME_FROM_ABOVE> |grep Status
  ```

## Deploy Sample App Locally

You will need two windows

1. Vault & Cello Service
1. Client commands, etc

### One Time Setup

- In window **#1**, ensure you have AWS credentials for the target account.

- Create the IAM role which will be used for the sample project.

  ```sh
  bash scripts/create_iam_role.sh
  ```

- Create a new postgres database. This can be done using the command:

  ```sh
  createdb cello
  ```

- Use `golang-migrate` to create the relevant tables and create a new user with read/write permissions. This can be done using the command:

  ```sh
  migrate -path scripts/db_migrations -database 'postgres://localhost:5432/cello?sslmode=disable' up
  ```

- Create the default workflow template in Argo.

  ```sh
  argo template create -n argo workflows/cello-single-step-vault-aws.yaml
  ```

### Start Vault & Cello Service

- In window **#1** first set the **CELLO_ADMIN_SECRET** to a 16
  character string, this will be used to authorize admin commands against
  the Cello service.

      ```sh
      export CELLO_ADMIN_SECRET=abcd1234abcd1234
      ```

- Start the Cello Service (includes vault)

  ```sh
  make ; make up
  ```

- To run in debug mode set log level DEBUG before running

  ```
  export CELLO_LOG_LEVEL=DEBUG
  make ; make up
  ```

### Create Cello Project And Target (One Time Setup)

- In window **#2**, ensure you have the **CELLO_ADMIN_SECRET**
  env set to the same value used above.

- Ensure your credentials are set for the **target account** and create your first
  project and target. This returns the **CELLO_USER_TOKEN** for the new project.

      ```sh
      bash scripts/create_project.sh https://github.com/cello-proj/cello.git
      ```

### Run Workflow

- Ensure the **CELLO_USER_TOKEN** for the project is specified

- CDK Example

  ```sh
  # CDK Example
  CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/cdk_manifest.yaml e3a419e69a5ae762862dc7cf382304a4e6cc2547 dev`

  # Get the status / logs
  ./build/cello get $CDK_WORKFLOW_NAME
  ./build/cello logs $CDK_WORKFLOW_NAME
  ```

- TERRAFORM Example

  ```sh
  # Terraform Example
  TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/terraform_manifest.yaml e3a419e69a5ae762862dc7cf382304a4e6cc2547 dev`

  # Get the status / logs
  ./build/cello get $TERRAFORM_WORKFLOW_NAME
  ./build/cello logs $TERRAFORM_WORKFLOW_NAME
  ```

