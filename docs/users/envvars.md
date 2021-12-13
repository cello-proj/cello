# Environment Variables

Argo CloudOps uses a number of environment variables for configuration. In addition to the table below, you can review the [start_local.sh](https://github.com/argoproj-labs/argo-cloudops/blob/main/scripts/start_local.sh) script for examples.

| Name                                       | Description                                                                                                                                    |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| ARGO_CLOUDOPS_ADMIN_SECRET                 | Secret for the Argo CloudOps API                                                                                                               |
| VAULT_ROLE                                 | Role for accessing Vault API                                                                                                                   |
| VAULT_SECRET                               | Secret for access Vault instance                                                                                                               |
| VAULT_ADDR                                 | Endpoint for the Vault instance                                                                                                                |
| ARGO_ADDR                                  | Argo Endpoint                                                                                                                                  |
| ARGO_CLOUDOPS_WORKFLOW_EXECUTION_NAMESPACE | Namespace to use to execute the deployments in Argo Workflows (Default: argo)                                                                  |
| ARGO_CLOUDOPS_CONFIG                       | File that contains argo cloudops command configuration. [Example](https://github.com/argoproj-labs/argo-cloudops/blob/main/argo-cloudops.yaml) |
| SSH_PEM_FILE                               | PEM file to use for GITHUB access authentication                                                                                               |
| ARGO_CLOUDOPS_GIT_AUTH_METHOD              | A value of SSH or HTTPS depending on which authentication method prefered.                                                                     |
| ARGO_CLOUDOPS_GIT_HTTPS_USER               | User name for GITHUB access authentication via HTTPS.                                                                                          |
| ARGO_CLOUDOPS_GIT_HTTPS_PASS               | Password for GITHUB access authentication via HTTPS.                                                                                           |
| ARGO_CLOUDOPS_DB_HOST                      | Database Host                                                                                                                                  |
| ARGO_CLOUDOPS_DB_USER                      | Database User                                                                                                                                  |
| ARGO_CLOUDOPS_DB_PASSWORD                  | Database Password                                                                                                                              |
| ARGO_CLOUDOPS_DB_NAME                      | Database name                                                                                                                                  |
| ARGO_CLOUDOPS_LOG_LEVEL                    | The configured log level for Argo CloudOps service (Default: Info)                                                                             |
| ARGO_CLOUDOPS_PORT                         | Port which the Argo CloudOps service listens (Default: 8443)                                                                                   |
| ARGO_CLOUDOPS_IMAGE_URIS                   | List of approved image URI patterns. See IsApprovedImageURI validation doc for examples                                                        |
