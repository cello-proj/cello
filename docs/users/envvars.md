# Environment Variables

Cello uses a number of environment variables for configuration. In addition to the table below, you can review the [start_local.sh](https://github.com/cello-proj/cello/blob/main/scripts/start_local.sh) script for examples.

| Name                                       | Description                                                                                                                         |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| CELLO_ADMIN_SECRET                 | Secret for the Cello API                                                                                                    |
| VAULT_ROLE                                 | Role for accessing Vault API                                                                                                        |
| VAULT_SECRET                               | Secret for access Vault instance                                                                                                    |
| VAULT_ADDR                                 | Endpoint for the Vault instance                                                                                                     |
| ARGO_ADDR                                  | Argo Endpoint                                                                                                                       |
| CELLO_WORKFLOW_EXECUTION_NAMESPACE | Namespace to use to execute the deployments in Argo Workflows (Default: argo)                                                       |
| CELLO_CONFIG                       | File that contains cello command configuration. [Example](https://github.com/cello-proj/cello/blob/main/cello.yaml)
| SSH_PEM_FILE                               | PEM file to use for GITHUB access authentication                                                                                    |
| CELLO_GIT_AUTH_METHOD              | A value of SSH or HTTPS depending on which authentication method prefered.                                                          |
| CELLO_GIT_HTTPS_USER               | User name for GITHUB access authentication via HTTPS.                                                                               |
| CELLO_GIT_HTTPS_PASS               | Password for GITHUB access authentication via HTTPS.                                                                                |
| CELLO_DB_HOST                      | Database Host                                                                                                                       |
| CELLO_DB_USER                      | Database User                                                                                                                       |
| CELLO_DB_PASSWORD                  | Database Password                                                                                                                   |
| CELLO_DB_NAME                      | Database name                                                                                                                       |
| CELLO_LOG_LEVEL                    | The configured log level for Cello service (Default: Info)                                                                  |
| CELLO_PORT                         | Port which the Cello service listens (Default: 8443)                                                                        |
| CELLO_IMAGE_URIS                   | List of approved image URI patterns. See IsApprovedImageURI validation doc for examples                                             |
