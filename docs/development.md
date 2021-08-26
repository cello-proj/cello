# Development Environment setup

## Pre-reqs

The below instructions assume Argo CloudOps is on your local OSX system
with Docker Desktop managing resource in AWS (region us-west-2) with credentials provided by Vault.

* Install [Docker Desktop](https://www.docker.com/products/docker-desktop), ensure kubernetes is running.

* Install [Argo Workflows](https://argoproj.github.io/argo-workflows/installation/)

* Install **Argo CLI** `brew install argo`

* Install **GoLang** `brew install golang`

* Install GoLint `go get -u golang.org/x/lint/golint` and ensure `$GOPATH` is in your `$PATH`.

* Install [PostgreSQL](https://www.postgresql.org/download/)

* Submit Argo Hello World workflow and record the **Name** from the output.

    ```
    argo submit -n argo https://raw.githubusercontent.com/argoproj/argo-workflows/master/examples/hello-world.yaml
    ```

* Ensure the workflow completes with Status **Succeeded**.

    ```
    argo get -n argo <UPDATE_WITH_NAME_FROM_ABOVE> |grep Status
    ```

* Install [Vault](https://www.vaultproject.io/downloads) for credential generation.

* Install [jq](https://stedolan.github.io/jq/) for json parsing.

* Install **npm** `brew install npm` (For CDK).

* Install [terraform](https://www.terraform.io/downloads.html).

## Deploy Sample App Locally

You will need two windows

1. Vault & Argo CloudOps Service
1. Client commands, etc

### One Time Setup

* In window **#1**, ensure you have AWS credentials for the target account.

* Create the IAM role which will be used for the sample project.

    ```sh
    bash scripts/create_iam_role.sh
    ```

* Create a new postgres database. This can be done using the command:

    ```sh
    createdb argocloudops
    ```

* Use the `createdbtables.sql` script to create the relevant tables and create a new user with read/write permissions. This can be done using the command:

    ```sh
    psql -d argocloudops -f scripts/createdbtables.sql
    ```

* Create the default workflow template in Argo.

    ```sh
    argo template create -n argo workflows/argo-cloudops-single-step-vault-aws.yaml
    ```

### Start Vault & Argo CloudOps Service

* In window **#1** first set the **ARGO_CLOUDOPS_ADMIN_SECRET** to a 16
character string, this will be used to authorize admin commands against
the Argo CloudOps service.

    ```sh
    export ARGO_CLOUDOPS_ADMIN_SECRET=abcd1234abcd1234
    ```

* Start the Argo CloudOps Service (includes vault)

    ```sh
    make ; make up
    ```

* To run in debug mode set log level DEBUG before running

    ```
    export ARGO_CLOUDOPS_LOG_LEVEL=DEBUG
    make ; make up
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
    CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/cdk_manifest.yaml 8bacf9cd5cf08c142fd5d29317a4d072bdd0800c dev`

    # Get the status / logs
    ./build/argo-cloudops get $CDK_WORKFLOW_NAME
    ./build/argo-cloudops logs $CDK_WORKFLOW_NAME
    ```

* TERRAFORM Example

    ```sh
    # Terraform Example
    TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/terraform_manifest.yaml 8bacf9cd5cf08c142fd5d29317a4d072bdd0800c dev`

    # Get the status / logs
    ./build/argo-cloudops get $TERRAFORM_WORKFLOW_NAME
    ./build/argo-cloudops logs $TERRAFORM_WORKFLOW_NAME
    ```
