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

| PostgreSQL         | DynamoDB pk              | DynamoDB sk        | Additional Attributes      |
| ------------------ | ------------------------ | ------------------ | -------------------------- |
| `projects.project` | `PROJECT#<project_name>` | `METADATA`         | `repository`               |
| `tokens.token_id`  | `PROJECT#<project_name>` | `TOKEN#<token_id>` | `created_at`, `expires_at` |

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

## Step 4: Fix Timestamp Format (Post-Migration)

> [!IMPORTANT]
> Due to a bug in the initial migration code, timestamps may have been loaded
> in PostgreSQL format (`2022-02-02 18:01:49.345261+00`) instead of the
> required ISO 8601 format (`2022-02-02T18:01:49.345261000Z`). This step
> corrects the timestamp format in DynamoDB.

### 4.1: Dump Current DynamoDB Data

First, dump the current data from DynamoDB to a JSON file for inspection and migration:

#### Command Syntax

```bash
./scripts/dump-ddb.py --table <table_name> --output <output_file> --region <region>
```

#### Parameters

- `--table`: Name of the DynamoDB table (e.g., `cello`)
- `--output`: Path to the output JSON file
- `--region`: AWS region where the DynamoDB table is located

#### Example Usage

```bash
# Dump all items from DynamoDB table
./scripts/dump-ddb.py \
  --table cello \
  --output ddb-dump.json \
  --region us-west-2
```

#### What the Script Does

- Scans the entire DynamoDB table
- Converts DynamoDB typed JSON (AttributeValue format) to plain Python dictionaries
- Outputs all items to a JSON array file

### 4.2: Migrate Timestamps (Dry Run)

Before making actual changes, run the migration script in dry-run mode to preview what will be updated:

#### Command Syntax

```bash
./scripts/migrate_ddb_timestamps.py --file <input_file> --table <table_name> --region <region> --dry-run
```

#### Parameters

- `--file`: Path to the JSON file created in Step 4.1
- `--table`: Name of the DynamoDB table (e.g., `cello`)
- `--region`: AWS region where the DynamoDB table is located
- `--dry-run`: Preview changes without making actual updates

#### Example Usage

```bash
# Preview timestamp migrations
./scripts/migrate_ddb_timestamps.py \
  --file ddb-dump.json \
  --table cello \
  --region us-west-2 \
  --dry-run
```

#### What the Script Does

- Loads records from the JSON dump file
- Identifies records with timestamps in PostgreSQL format (ending in `+00`)
- Converts timestamps to ISO 8601 format with nanosecond precision (ending in `Z`)
- Shows what would be updated without making changes
- Skips:
  - METADATA records (no timestamps to migrate)
  - Records already in ISO 8601 format (ending in `Z`)
  - Records with timestamps not in PostgreSQL format

#### Expected Output

The script will display each record that needs updating:

```
[PROJECT#my-project][TOKEN#tkn-123]:
  created_at: 2022-02-02 18:01:49.345261+00 → 2022-02-02T18:01:49.345261000Z
  expires_at: 2023-02-02 18:01:49.345261+00 → 2023-02-02T18:01:49.345261000Z
  DRY RUN - Would update
```

### 4.3: Migrate Timestamps (Live Update)

After reviewing the dry-run output and confirming the changes look correct, run the script without the `--dry-run` flag to perform the actual updates:

#### Command Syntax

```bash
./scripts/migrate_ddb_timestamps.py --file <input_file> --table <table_name> --region <region>
```

#### Example Usage

```bash
# Perform actual timestamp migration
./scripts/migrate_ddb_timestamps.py \
  --file ddb-dump.json \
  --table cello \
  --region us-west-2
```

#### What the Script Does

- Processes each record that needs timestamp conversion
- Updates DynamoDB records using `update_item` with conditional expressions
- Provides detailed output for each update
- Reports summary statistics at completion

#### Expected Output

```
Configuration:
  Input file: ddb-dump.json
  DynamoDB table: cello
  Region: us-west-2
  Dry run: False

Loading records from ddb-dump.json...
Loaded 150 total records

Processing records:
--------------------------------------------------------------------------------
[PROJECT#my-project][TOKEN#tkn-123]:
  created_at: 2022-02-02 18:01:49.345261+00 → 2022-02-02T18:01:49.345261000Z
  expires_at: 2023-02-02 18:01:49.345261+00 → 2023-02-02T18:01:49.345261000Z
  ✓ Updated

--------------------------------------------------------------------------------
Summary:
  Total records: 150
  Skipped: 100
  Processed: 50
  Successful: 50
  Failed: 0
```

### Timestamp Format Details

The migration converts timestamps from PostgreSQL format to ISO 8601/RFC3339 format:

- **Before**: `2022-02-02 18:01:49.345261+00`
  - Space between date and time
  - Ends with `+00` (timezone offset)
  - Microsecond precision (6 digits)

- **After**: `2022-02-02T18:01:49.345261000Z`
  - `T` separator between date and time
  - Ends with `Z` (UTC timezone)
  - Nanosecond precision (9 digits, zero-padded)

This format is compatible with Go's `time.RFC3339Nano` constant and is the standard format used throughout the Cello service.

