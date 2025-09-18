# PostgreSQL to DynamoDB Migration Guide

This document provides step-by-step instructions for migrating Cello data from PostgreSQL to DynamoDB using the provided migration scripts. The migration follows the [DynamoDB Single-Table Design](dynamodb_single_table_design.md) specification.

## Overview

The migration process consists of two main phases:

1. **Data Extraction**: Dump data from PostgreSQL to JSON files using `migrate_dump_postgres.py`
2. **Data Loading**: Load the JSON data into DynamoDB using `migrate_load_dynamodb.py`

## Prerequisites

### Requirements

- Python 3.12 or higher
- [`uv`](https://github.com/astral-sh/uv) package manager installed
- `psql` command-line client (for PostgreSQL access)
- Access to the source PostgreSQL database
- AWS credentials configured for DynamoDB access
- DynamoDB table created according to the single-table design

### AWS Configuration
Ensure your AWS credentials are properly configured with permissions to:
- Describe DynamoDB tables
- Write items to the target DynamoDB table

## Step 1: Extract Data from PostgreSQL

### Command Syntax
```bash
uv run scripts/migrate_dump_postgres.py <db_host[:port]> <db_user> <db_name> <output_file>
```

### Parameters
- `db_host[:port]`: PostgreSQL server hostname or IP address, optionally with port (e.g., `localhost:5433`)
- `db_user`: PostgreSQL username
- `db_name`: PostgreSQL database name
- `output_file`: Path to the output JSON file

### Host and Port Format
The `db_host` parameter supports flexible host:port parsing:

- `localhost` - Uses default PostgreSQL port (5432)
- `localhost:5433` - Uses custom port 5433
- `db.example.com:5432` - Remote server with custom port
- `192.168.1.100:5432` - IP address with custom port

**Note**: The script automatically detects the `:` separator and parses the port number. If no port is specified, it defaults to the standard PostgreSQL port 5432.

### Password Handling
The script provides two options for the DB password:

1. **Environment variable**: Set `PGPASSWORD=mypassword` before running the script
2. **Interactive prompt**: If `PGPASSWORD` is not set, the script will prompt for the password 

**Security Note**: The `PGPASSWORD` environment variable is the standard PostgreSQL authentication method and is recognized by the `psql` command. The script only sets this variable for the subprocess execution and doesn't persist it in the parent environment.

### Example Usage
```bash
# Basic usage with default port (5432) - password will be prompted
uv run scripts/migrate_dump_postgres.py \
  localhost \
  cello_user \
  cello_db \
  migration_data.json

# With custom port using host:port format
uv run scripts/migrate_dump_postgres.py \
  localhost:5433 \
  cello_user \
  cello_db \
  migration_data.json

# With remote server and custom port
uv run scripts/migrate_dump_postgres.py \
  db.example.com:5432 \
  cello_user \
  cello_db \
  migration_data.json

# Using PGPASSWORD environment variable for password
PGPASSWORD=mypassword uv run scripts/migrate_dump_postgres.py \
  localhost \
  cello_user \
  cello_db \
  migration_data.json

# With IP address and custom port
uv run scripts/migrate_dump_postgres.py \
  192.168.1.100:5432 \
  cello_user \
  cello_db \
  migration_data.json
```

### What the Script Does
The script grabs all the relevant data from PostgreSQL and dumps it into a JSON file.

The script extracts:
- **Projects**: `project` and `repository` fields
- **Tokens**: `token_id`, `created_at`, `project`, and `expires_at` fields

### Output Structure
The script generates a JSON file with the following structure:

```json
{
  "project_name": {
    "repository": "https://github.com/example/project",
    "tokens": [
      {
        "token_id": "tkn-123",
        "created_at": "2023-06-15T12:00:00Z",
        "expires_at": "2023-12-15T12:00:00Z"
      }
    ]
  }
}
```

## Step 2: Load Data into DynamoDB

### Command Syntax
```bash
uv run scripts/migrate_load_dynamodb.py <data_file> <table_name> <region>
```

### Parameters
- `data_file`: Path to the JSON file created in Step 1
- `table_name`: Name of the DynamoDB table (e.g., `cello`)
- `region`: AWS region where the DynamoDB table is located

### Example Usage
```bash
# Load data into DynamoDB
uv run scripts/migrate_load_dynamodb.py \
  migration_data.json \
  cello \
  us-west-2
```

### What the Script Does
1. Loads the JSON data from the file
2. Connects to the specified DynamoDB table
3. For each project:
   - Creates a project metadata item with `pk="PROJECT#<project_name>"` and `sk="METADATA"`
   - Creates token items with `pk="PROJECT#<project_name>`" and `sk="TOKEN#<token_id>"`
4. Uses batch writes for efficiency
5. Implements retry logic for failed writes
6. Skips items that already exist (idempotent operation)

### Data Mapping
The script maps PostgreSQL data to DynamoDB items as follows:

| PostgreSQL | DynamoDB pk | DynamoDB sk | Additional Attributes |
|------------|-------------|-------------|----------------------|
| `projects.project` | `PROJECT#<project_name>` | `METADATA` | `repository` |
| `tokens.token_id` | `PROJECT#<project_name>` | `TOKEN#<token_id>` | `created_at`, `expires_at` |

## Step 3: Verify Migration

After the migration completes, verify the data integrity:

### Check Migration Summary
The script provides a summary of:
- Number of projects processed
- Total items loaded
- Total items skipped (already existed)
- Overall success status

### Verify Data in DynamoDB
```bash
# Check project metadata
aws dynamodb get-item \
  --table-name cello \
  --key '{"pk":{"S":"PROJECT#your_project"},"sk":{"S":"METADATA"}}'

# Check tokens for a project
aws dynamodb query \
  --table-name cello \
  --key-condition-expression "pk = :pk AND begins_with(sk, :sk)" \
  --expression-attribute-values '{":pk":{"S":"PROJECT#your_project"},"sk":{"S":"TOKEN#"}}'
```

### Compare with Source Data
- Verify project count matches
- Verify token count per project matches
- Spot-check individual records for accuracy