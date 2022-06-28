# API

## Create Project

POST /projects

Request Body

```json
{
  "name": "project1",
  "repository": "git@github.com:myorg/myrepo.git"
}
```

Response Body

```json
{
  "token": "abcd-1234"
}
```

## Get Project

GET /projects/<project_name>

Response Body

```json
{
  "name": "myproject",
  "repository": "git@github.com:myorg/myrepo.git"
}
```

## Delete Project

DELETE /projects/<project_name>

Projects can only be deleted if they have no targets

Response Body

```
```

## Create Target

POST /projects/<project_name>/targets

Request Body

```json
{
  "name": "target1",
  "type": "aws_account",
  "properties": {
    "credential_type": "assumed_role",
    "policy_arns": [
      "arn:aws:iam::aws:policy/AmazonS3FullAccess",
      "arn:aws:iam::aws:policy/AmazonSNSFullAccess",
      "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
      "arn:aws:iam::aws:policy/AWSCloudFormationFullAccess"
    ],
    "policy_document": "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"s3:ListBuckets\", \"Resource\": \"*\" } ] }",
    "role_arn": "arn:aws:iam::<ACCOUNT_ID>:role/<ROLE_NAME>"
  }
}
```

Note: `role_arn` will be assumed as the target by vault. Vault's IAM
credentials must be a principle authorized to assume this role. The
`policy_arns` and `policy_document` will be applied at role assumption time to
scope down permissions. Today only type is only `aws_account` and
`credential_type` is only assumed role.

Response Body

```json
{}
```

# List Targets

GET /projects/<project_name>/targets

Response Body

```json
["target1", "target2"]
```

# Get Target

GET /projects/<project_name>/targets/<target_name>

Response Body

```json
{
  "name": "target1",
  "type": "aws_account",
  "properties": {
    "credential_type": "assumed_role",
    "policy_arns": [
      "arn:aws:iam::aws:policy/AmazonS3FullAccess",
      "arn:aws:iam::aws:policy/AmazonSNSFullAccess",
      "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
      "arn:aws:iam::aws:policy/AWSCloudFormationFullAccess"
    ],
    "policy_document": "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"s3:ListBuckets\", \"Resource\": \"*\" } ] }",
    "role_arn": "arn:aws:iam::123456789012:role/CelloSampleRole"
  }
}
```

## Update Target

PATCH /projects/<project_name>/targets/<target_name>

Request Body

```json
{
  "properties": {
    "policy_arns": [
      "arn:aws:iam::aws:policy/AmazonS3FullAccess",
      "arn:aws:iam::aws:policy/AmazonSNSFullAccess",
      "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
      "arn:aws:iam::aws:policy/AWSCloudFormationFullAccess"
    ],
    "policy_document": "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"s3:ListBuckets\", \"Resource\": \"*\" } ] }",
    "role_arn": "arn:aws:iam::<ACCOUNT_ID>:role/<ROLE_NAME>"
  }
}
```

Note: Target properties that are provided will be updated with the new values provided.
Properties that are not provided in the PATCH request will remain with their current values.
`credential_type` cannot be updated

Response Body

```json
{
  "name": "target1",
  "type": "aws_account",
  "properties": {
    "credential_type": "assumed_role",
    "policy_arns": [
      "arn:aws:iam::aws:policy/AmazonS3FullAccess",
      "arn:aws:iam::aws:policy/AmazonSNSFullAccess",
      "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
      "arn:aws:iam::aws:policy/AWSCloudFormationFullAccess"
    ],
    "policy_document": "{ \"Version\": \"2012-10-17\", \"Statement\": [ { \"Effect\": \"Allow\", \"Action\": \"s3:ListBuckets\", \"Resource\": \"*\" } ] }",
    "role_arn": "arn:aws:iam::123456789012:role/CelloSampleRole"
  }
}
```

## Delete Target

DELETE /projects/<project_name>/targets/<target_name>

Response Body

```
```

## List Project Tokens

GET /projects/<project_name>/tokens

Response Body

```json
[
  {
    "created_at": "2022-06-21T14:56:10.341066-07:00",
    "token_id": "ghi789"
  },
  {
    "created_at": "2022-06-21T14:43:16.172896-07:00",
    "token_id": "def456"
  },
]
```

## Create Workflow

POST /workflows

Request Body

```json
{
  "arguments": {
    "execute": [
      "-auto-approve",
      "-no-color"
    ],
    "init": [
      "-no-color"
    ]
  },
  "environment_variables": {
    "AWS_REGION": "us-west-2",
    "CODE_URI": "s3://cello-cet-dev/terraform-example.zip",
    "VAULT_ADDR": "http://docker.for.mac.localhost:8200"
  },
  "framework": "terraform",
  "parameters": {
    "execute_container_image_uri": "a80addc4/cello-terraform:0.14.5"
  },
  "project_name": "project1",
  "target_name": "target1",
  "type": "sync",
  "workflow_template_name": "cello-single-step-vault-aws"
}
```

Note: Arguments will be concatenated with spaces before appended to the command.

Response Body

```json
{
  "workflow_name": "abcd"
}
```

## Perform Target Operations From Git Manifest

POST /projects/<project_name>/targets/<target_name>/operations

Request Body

```json
{
  "sha": "1234abdc5678efgh9012ijkl3456mnop7890qrst",
  "path": "path/to/manifest.yaml"
}
```

Response Body

```json
{
  "workflow_name": "abcd"
}
```

## Get Workflow

GET /workflows/<workflow_name>

Response Body

```json
{
  "name":"workflow1",
  "status":"failed",
  "created":"1618515183",
  "finished":"1618515193"
}
```

## Get Workflow Logs

GET /workflows/<workflow_name>/logs

Response Body

```json
{
  "logs": [
    "Log line 1",
    "Log line 2"
  ]
}
```

## Get Workflow Logstream

GET /workflows/<workflow_name>/logstream

Response Body

```text
  Log line 1
  Log line 2
```

# List Project / Target Workflows

GET /projects/<project_name>/targets/<target_name>/workflows

Response Body

```json

[
  {"name":"workflow1","status":"failed","created":"1618515183","finished":"1618515193"},
  {"name":"workflow2","status":"failed","created":"1618512676","finished":"1618512686"}
]
```
