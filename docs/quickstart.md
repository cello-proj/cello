# Quickstart

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

* Create a new database and use the `createdb.sql` script to create the relevant tables

* Create a database user with read/write permissions to the created database and set the `ARGO_CLOUDOPS_DB_NAME`, `ARGO_CLOUDOPS_DB_USER`, and `ARGO_CLOUDOPS_DB_PASSWORD` variables as needed. The default values if the variables are not set are `argocloudops`, `argoco`, and `1234` respectively

* Create an S3 bucket (change the bucket name below) and set it as **ARGO_CLOUDOPS_BUILD_BUCKET** environment variable:

    ```sh
    export ARGO_CLOUDOPS_BUILD_BUCKET=<UPDATE_ME>
    aws s3 mb s3://${ARGO_CLOUDOPS_BUILD_BUCKET} --region us-west-2
    ```

* Create the IAM role which will be used for the sample project.

    ```sh
    bash scripts/create_iam_role.sh
    ```

* Build and upload the framework code to S3.

    ```sh
    cd ./examples ; make ; cd -
    ```

* Create a fork of the [example repository](https://github.com/Acepie/argo-cloudops-example) and update the `CODE_URI` in the `manifest.yaml` file to use the correct build bucket based on the `ARGO_CLOUDOPS_BUILD_BUCKET` variable set

* Create the default workflow template in Argo.

    ```sh
    argo template create -n argo workflows/argo-cloudops-single-step-vault-aws.yaml
    ```

* Create the terraform state bucket (can be the same bucket as the ARGO_CLOUDOPS_BUILD_BUCKET for dev).

    ```sh
    export TERRAFORM_STATE_BUCKET=<UPDATE_ME>
    aws s3 mb "s3://${TERRAFORM_STATE_BUCKET}" --region us-west-2
    ```

* Update `./examples/app-terraform/main.tf` with your bucket name and region in
the 'backend' section and init terraform.

    ```sh
    terraform init ./examples/app-terraform/
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
    export ARGO_CLOUD_OPS_LOG_LEVEL=DEBUG
    make ; make up
    ```

### Create Argo CloudOps Project And Target (One Time Setup)

* In window **#2**, ensure you have the **ARGO_CLOUDOPS_ADMIN_SECRET**
env set to the same value used above.

* Ensure your credentials are set for the **target account** and create your first
project and target. This returns the **ARGO_CLOUDOPS_USER_TOKEN** for the new project. For the git repo, use the example fork that was made earlier

    ```sh
    bash scripts/create_project.sh git@github.com:Acepie/argo-cloudops-example.git
    ```

### Run Workflow

* Ensure the **ARGO_CLOUDOPS_USER_TOKEN** for the project is specified. The second argument for the bash commands below should be the commit sha for the commit on your fork that has your manifest

* CDK Example

    ```sh
    # CDK Example
    CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/cdk_manifest.yaml 0cb2797bfae9d0d18f6ab22c3e1fde5ac170be5e`

    # Terraform Example
    TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/terraform_manifest.yaml 0cb2797bfae9d0d18f6ab22c3e1fde5ac170be5e`

    # Get the status / logs
    ./build/argo-cloudops get $TERRAFORM_WORKFLOW_NAME
    ./build/argo-cloudops logs $TERRAFORM_WORKFLOW_NAME
    ```
