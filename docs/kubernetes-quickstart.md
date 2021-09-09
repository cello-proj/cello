# Quickstart

## Pre-reqs

* Install [Docker Desktop](https://www.docker.com/products/docker-desktop), ensure kubernetes is running.

* Install [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)

* Install **Argo CLI** `brew install argo`

## Deploy Sample App Locally

You will need two windows

1. Vault & Argo CloudOps Service
1. Client commands, etc


### Start Vault & Argo CloudOps Service

* In window **#1**, ensure you have AWS credentials for the target account.

* Build the docker image for the Argo CloudOps Service

    ```sh
    docker build --pull --rm -f "Dockerfile" -t argocloudops:latest "."
    ```

* Apply the kubernetes manifest in the scripts folder. This will start up Argo Workflows, Vault, Postgres, and the Argo CloudOps Service

    ```sh
    kubectl apply -f ./scripts/quickstart_manifest.yaml
    ```

* Initialize postgres with the necessary database and tables

    ```sh
    export POSTGRES_POD="$(kubectl get pods --no-headers -o custom-columns=":metadata.name" | grep postgres)"
    kubectl cp ./scripts/createdbtables.sql $POSTGRES_POD:./createdbtables.sql
    kubectl exec $POSTGRES_POD -- createdb argocloudops -U postgres
    kubectl exec $POSTGRES_POD -- psql -U postgres -d argocloudops -f ./createdbtables.sql
    ```

* Create the sample argo workflow template

    ```sh
    argo template create -n argo workflows/argo-cloudops-single-step-vault-aws.yaml
    ```

* Retrieve the AWS credentials for your target account and send them to vault

    ```sh
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
    ```

* Forward a port to the Argo CloudOps Service to run commands against

    ```sh
    export ACO_POD="$(kubectl get pods --no-headers -o custom-columns=":metadata.name" | grep argocloudops)"
    kubectl port-forward $ACO_POD 8443:8443
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
    CDK_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/kube_cdk_manifest.yaml <TODO add SHA>`

    # Get the status / logs
    ./quickstart/argo-cloudops get $CDK_WORKFLOW_NAME
    ./quickstart/argo-cloudops logs $CDK_WORKFLOW_NAME
    ```

* TERRAFORM Example

    ```sh
    # Terraform Example
    TERRAFORM_WORKFLOW_NAME=`bash scripts/run_gitops_example.sh manifests/kube_terraform_manifest.yaml <TODO add SHA>`

    # Get the status / logs
    ./quickstart/argo-cloudops get $TERRAFORM_WORKFLOW_NAME
    ./quickstart/argo-cloudops logs $TERRAFORM_WORKFLOW_NAME
    ```