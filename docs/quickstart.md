# Quickstart

## Pre-reqs

* Install [Docker Desktop](https://www.docker.com/products/docker-desktop), ensure kubernetes is running.

* Install [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)

* Install **Argo CLI** `brew install argo`

* Install [jq](https://stedolan.github.io/jq/) for json parsing.

## Deploy Sample App Locally

You will need two windows

1. Vault & Argo CloudOps Service
1. Client commands, etc


### Start Vault & Argo CloudOps Service

* In window **#1**, ensure you have AWS credentials for the target account.

* Start the Argo CloudOps Service (includes workflows, vault, and postgres)

    ```sh
    bash scripts/quickstart_run.sh
    ```

### Create Argo CloudOps Project And Target (One Time Setup)

* In window **#2**, ensure you have the **ARGO_CLOUDOPS_ADMIN_SECRET**
env set to `abcd1234abcd1234`.

    ```sh
    export ARGO_CLOUDOPS_ADMIN_SECRET=abcd1234abcd1234
    ```

* Ensure your credentials are set for the **target account** and create your first
project and target. This returns the **ARGO_CLOUDOPS_USER_TOKEN** for the new project.

    ```sh
    bash scripts/create_project.sh https://github.com/argoproj-labs/argo-cloudops.git
    ```

### Run Workflow

* Ensure the **ARGO_CLOUDOPS_USER_TOKEN** for the project is specified

* CDK Example

    ```sh
    # CDK Example
    CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/kube_cdk_manifest.yaml 5b40793bded1030d8a17d6ddd050ee1ef060f8cc`

    # Get the status / logs
    ./quickstart/argo-cloudops get $CDK_WORKFLOW_NAME
    ./quickstart/argo-cloudops logs $CDK_WORKFLOW_NAME
    ```

* TERRAFORM Example

    ```sh
    # Terraform Example
    TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/kube_terraform_manifest.yaml 5b40793bded1030d8a17d6ddd050ee1ef060f8cc`

    # Get the status / logs
    ./quickstart/argo-cloudops get $TERRAFORM_WORKFLOW_NAME
    ./quickstart/argo-cloudops logs $TERRAFORM_WORKFLOW_NAME
    ```
