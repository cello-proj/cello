# Core Concepts

This page serves as additional information in addition to the [Cello Architecture](../architecture.md).

## Project

A project is a logical collection of all deployment targets.

### Properties

| Name       | Description                                                                                                                                                                                  |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name       | name for the project                                                                                                                                                                         |
| repository | link to the github repository with [all project manifests](https://github.com/cello-proj/cello/blob/main/manifests/cdk_manifest.yaml). Should match the auth method being used (HTTPS, SSH). |

## Target

A target represents a unique deployment for a project. It contains information related to cloud account access mechanism & policies for scoping permissions. Currently the only type of cloud account supported is AWS.

### Properties

| Name            | Description                                                            |
| --------------- | ---------------------------------------------------------------------- |
| credential_type | the type of credential mechanism to use. Currently only "assumed_role" |
| role_arn        | the role that the service assumes                                      |
| policy_arns     | A list of AWS policy ARNs to use for permissions scope limiting        |
| policy_document | An inline document to scope down permissions                           |
