# Quickstart

## Pre-reqs

* Install [Docker Desktop](https://www.docker.com/products/docker-desktop), ensure kubernetes is running.

* Install [Argo Workflows](https://argoproj.github.io/argo-workflows/installation/)

* Install **Argo CLI** `brew install argo`

* Install [PostgreSQL](https://www.postgresql.org/download/)

* Install [Vault](https://www.vaultproject.io/downloads) for credential generation.

* Install [jq](https://stedolan.github.io/jq/) for json parsing.

## Deploy Sample App Locally

You will need two windows

1. Vault & Argo CloudOps Service
1. Client commands, etc


### Start Vault & Argo CloudOps Service

* In window **#1**, ensure you have AWS credentials for the target account.

* In window **#1** first set the **ARGO_CLOUDOPS_ADMIN_SECRET** to a 16
character string, this will be used to authorize admin commands against
the Argo CloudOps service.

    ```sh
    export ARGO_CLOUDOPS_ADMIN_SECRET=abcd1234abcd1234
    ```

* Start the Argo CloudOps Service (includes vault)

    ```sh
    ./scripts/quickstart_run.sh
    ```

* To run in debug mode set log level DEBUG before running

    ```
    export ARGO_CLOUDOPS_LOG_LEVEL=DEBUG
    ./scripts/quickstart_run.sh
    ```

### Create Argo CloudOps Project And Target (One Time Setup)

* In window **#2**, ensure you have the **ARGO_CLOUDOPS_ADMIN_SECRET**
env set to the same value used above.

* Ensure your credentials are set for the **target account** and create your first
project and target. This returns the **ARGO_CLOUDOPS_USER_TOKEN** for the new project.

    ```sh
    bash scripts/create_project.sh https://github.com/Acepie/argo-cloudops-example.git
    ```

### Run Workflow

* Ensure the **ARGO_CLOUDOPS_USER_TOKEN** for the project is specified

* CDK Example

    ```sh
    # CDK Example
    CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/cdk_manifest.yaml 8bacf9cd5cf08c142fd5d29317a4d072bdd0800c`

    # Get the status / logs
    ./quickstart/argo-cloudops get $CDK_WORKFLOW_NAME
    ./quickstart/argo-cloudops logs $CDK_WORKFLOW_NAME
    ```

* TERRAFORM Example

    ```sh
    # Terraform Example
    TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/terraform_manifest.yaml 8bacf9cd5cf08c142fd5d29317a4d072bdd0800c`

    # Get the status / logs
    ./quickstart/argo-cloudops get $TERRAFORM_WORKFLOW_NAME
    ./quickstart/argo-cloudops logs $TERRAFORM_WORKFLOW_NAME
    ```
