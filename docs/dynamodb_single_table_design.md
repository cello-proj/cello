# DynamoDB Single-Table Schema Proposal

This document describes a single-table DynamoDB design that captures the functionality of the PostgreSQL schema, including:

- Projects
- Tokens
- Targets (tbd)
- Dynamic TargetProperties (tbd)

## Table Name

Use a single DynamoDB table named `cello`.

## Primary Keys

- **Partition Key (pk):** Project identifier (string)
- **Sort Key (sk):** Item type and identifier (string)

## Entities and Item Structures

### 1. Project Items

• **pk**: `"PROJECT#<project_name>"`
• **sk**: `"META"`
• **Additional Attributes**:

- `repository` (string)

Example:

```json
{
  "pk": "PROJECT#myproj",
  "sk": "META",
  "repository": "https://github.com/example/myproj"
}
```

### 2. Token Items

• **pk**: `"PROJECT#<project_name>"`
• **sk**: `"TOKEN#<token_id>"`
• **Additional Attributes**:

- `token_id` (string)
- `created_at` (ISO-8601 date/time string)
- `expires_at` (ISO-8601 date/time string)
- `hashed_token` (string containing the hashed token value) (tbd - we are not implementing this right now. this is not currently stored; would be additional functionality to implement)

Example:

```json
{
  "pk": "PROJECT#myproj",
  "sk": "TOKEN#tkn-123",
  "token_id": "tkn-123",
  "created_at": "2023-06-15T12:00:00Z",
  "expires_at": "2023-12-15T12:00:00Z",
  "hashed_token": "<hashed-token-string>"
}
```

### 3. Target Items (tbd)

Each project can reference multiple Targets (see `Target` and `TargetProperties` in internal/types/types.go). We'll store each Target as one item.

• **pk**: `"PROJECT#<project_name>"`
• **sk**: `"TARGET#<target_name>"`
• **Additional Attributes**:

- `name`: The target name
- `type`: e.g., `"aws_account"`
- `properties`: A JSON-serialized representation of target-specific fields. Instead of storing each credential detail in separate columns (e.g., `credential_type`, `policy_arns`, etc.), the entire structure is persisted as a single JSON string or map. The application can parse this JSON to retrieve the relevant fields.

Example:

```json
{
  "pk": "PROJECT#myproj",
  "sk": "TARGET#mytarget",
  "name": "mytarget",
  "type": "aws_account",
  "properties": {
    "any_field_you_need": "some value",
    "more_nesting_here": {
      "subfield": "value"
    }
  }
}
```

Storing `properties` as a map (or JSON string) allows for easy extension of `TargetProperties` without changing the table schema.

## Access Patterns

1. **Get a Single Project**

   - `pk = "PROJECT#<project_name>"`, `sk = "META"`

2. **List All Tokens for a Project**

   - Query by `pk = "PROJECT#<project_name>"`
   - Filter or iterate over items where `sk` begins with `"TOKEN#"`

3. **Get a Single Token**

   - `pk = "PROJECT#<project_name>"`, `sk = "TOKEN#<token_id>"`

4. **Authenticate/Authorize (Token)** (tbd)

   - Retrieve the token item using `pk = "PROJECT#<project_name>"` and `sk = "TOKEN#<token_id>"`.
   - Compare `hashed_token` from the item with the hashed token in the request.

5. **List All Targets for a Project** (tbd)

   - Query by `pk = "PROJECT#<project_name>"`
   - Filter or iterate over items where `sk` begins with `"TARGET#"`.

6. **Get/Add/Update a Single Target** (tbd)

   - **Get**: `pk = "PROJECT#<project_name>"`, `sk = "TARGET#<target_name>"`
   - **Add/Update**: Put a new item (or update existing) with the same key: `pk = "PROJECT#<project_name>"`, `sk = "TARGET#<target_name>"`, along with attributes for `name`, `type`, and `properties`.

7. **Delete a Target** (tbd)
   - Use the same key (`pk` + `sk`).
   - `pk = "PROJECT#<project_name>"`, `sk = "TARGET#<target_name>"`.
   - Perform a delete operation.

## Data Access & Integrity

### Project Deletion and Cleanup

Deleting a Project is straightforward with the single-table design. The implementation:

1. Queries all items under the project's partition (`pk = "PROJECT#<project_name>"`)
2. Handles pagination to retrieve all items across multiple pages
3. Deletes all items in batches of 25 (DynamoDB's BatchWriteItem limit) with retry logic

This approach is efficient and leverages DynamoDB's single-table design where all related items (project metadata, tokens, targets) share the same partition key.

### Foreign Keys

DynamoDB does not enforce foreign keys. This single-table design keeps related items (Projects, Tokens, Targets) together in the same partition, but your application is responsible for implementing any referential integrity, such as cleaning up tokens and targets on project deletion.

---

## CloudFormation Resource

Below is an example CloudFormation YAML snippet that creates a DynamoDB table matching our single-table design. Since the current access patterns are fully supported by our primary key structure, no additional Global Secondary Index (GSI) is strictly required. Adjust as needed for your environment or capacity settings.

```yaml
Resources:
  CelloTable:
    Type: AWS::DynamoDB::Table
    Properties:
      TableName: cello
      BillingMode: PAY_PER_REQUEST
      AttributeDefinitions:
        - AttributeName: pk
          AttributeType: S
        - AttributeName: sk
          AttributeType: S
      KeySchema:
        - AttributeName: pk
          KeyType: HASH
        - AttributeName: sk
          KeyType: RANGE
      # No additional indexes required for our known access patterns
      # GlobalSecondaryIndexes: []

If future use cases require different query patterns (e.g., searching by token ID alone), then a new GSI or other approach might become necessary. However, given the patterns described above, this single-table design does not need extra indexes.

## Summary
The single-table approach handles Projects, Tokens, and Targets with flexible JSON for target properties, no additional indexes needed. This section provides an example CloudFormation template to implement the core table and primary key structure.
```
