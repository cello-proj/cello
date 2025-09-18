# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "boto3",
# ]
# ///
"""
PostgreSQL to DynamoDB Migration - Data Load Script

This script reads JSON files dumped from PostgreSQL and loads the data
into DynamoDB using the single-table design.

Usage:
    uv run scripts/migrate_load_dynamodb.py <data_file> <table_name> <region>
"""

import sys
import json
import boto3
from botocore.exceptions import ClientError, NoCredentialsError
from pathlib import Path
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


def migrate_project(table, project_name, project_data):
    """Migrate a single project and its tokens to DynamoDB."""
    logger.info(f"Migrating project: {project_name}")

    processed = 0
    skipped = 0

    try:
        # Validate project data structure
        if "repository" not in project_data:
            raise ValueError(f"Project {project_name} missing 'repository' field")
        if "tokens" not in project_data:
            raise ValueError(f"Project {project_name} missing 'tokens' field")
        if not isinstance(project_data["tokens"], list):
            raise ValueError(f"Project {project_name} 'tokens' field is not a list")

        # Add project metadata item
        project_item = {
            "pk": f"PROJECT#{project_name}",
            "sk": "METADATA",
            "repository": project_data["repository"],
        }

        # Write project item
        try:
            table.put_item(Item=project_item, ConditionExpression="attribute_not_exists(pk) AND attribute_not_exists(sk)")
            processed += 1
        except ClientError as e:
            if e.response["Error"]["Code"] == "ConditionalCheckFailedException":
                logger.info(f"Project {project_name} already exists, skipping")
                skipped += 1
            else:
                raise

        # Add token items
        for i, token in enumerate(project_data["tokens"]):
            # Validate token data structure
            if not isinstance(token, dict):
                logger.warning(f"Project {project_name} token {i} is not a dictionary, skipping")
                continue
            if "token_id" not in token:
                logger.warning(f"Project {project_name} token {i} missing 'token_id' field, skipping")
                continue
            if "created_at" not in token:
                logger.warning(f"Project {project_name} token {i} missing 'created_at' field, skipping")
                continue
            if "expires_at" not in token:
                logger.warning(f"Project {project_name} token {i} missing 'expires_at' field, skipping")
                continue

            token_item = {
                "pk": f"PROJECT#{project_name}",
                "sk": f"TOKEN#{token['token_id']}",
                "created_at": token["created_at"],
                "expires_at": token["expires_at"],
            }

            try:
                table.put_item(Item=token_item, ConditionExpression="attribute_not_exists(pk) AND attribute_not_exists(sk)")
                processed += 1
            except ClientError as e:
                if e.response["Error"]["Code"] == "ConditionalCheckFailedException":
                    logger.info(f"Token {token['token_id']} already exists, skipping")
                    skipped += 1
                else:
                    raise

    except Exception as e:
        logger.error(f"Error writing items for project {project_name}: {e}")
        raise

    logger.info(
        f"✓ Project {project_name}: {processed} items loaded, {skipped} skipped"
    )
    return processed, skipped


def main():
    if len(sys.argv) != 4:
        print(
            "Usage: uv run scripts/migrate_load_dynamodb.py <data_file> <table_name> <region>"
        )
        sys.exit(1)

    data_file = sys.argv[1]
    table_name = sys.argv[2]
    region = sys.argv[3]

    # Check if data file exists
    if not Path(data_file).exists():
        logger.error(f"Data file not found: {data_file}")
        sys.exit(1)

    # Load JSON data
    try:
        with open(data_file, "r") as f:
            migration_data = json.load(f)
        logger.info(f"✓ Loaded migration data: {len(migration_data)} projects")
    except json.JSONDecodeError as e:
        logger.error(f"Error parsing JSON file: {e}")
        sys.exit(1)

    # Initialize DynamoDB resource (higher-level interface)
    try:
        dynamodb = boto3.resource("dynamodb", region_name=region)
        table = dynamodb.Table(table_name)
        # Test connection by describing the table
        table.load()
        logger.info(f"✓ Connected to DynamoDB table: {table_name}")
    except NoCredentialsError:
        logger.error(
            "AWS credentials not found. Please configure your AWS credentials."
        )
        sys.exit(1)
    except ClientError as e:
        if e.response["Error"]["Code"] == "ResourceNotFoundException":
            logger.error(
                f"DynamoDB table '{table_name}' not found in region '{region}'"
            )
        else:
            logger.error(f"Error connecting to DynamoDB: {e}")
        sys.exit(1)

    # Migrate data
    logger.info("Starting data migration...")

    total_processed = 0
    total_skipped = 0
    projects_processed = 0
    projects_failed = 0
    total_projects = len(migration_data)

    for project_name, project_data in migration_data.items():
        logger.info(f"Processing project {projects_processed + projects_failed + 1}/{total_projects}: {project_name}")
        try:
            processed, skipped = migrate_project(
                table, project_name, project_data
            )
            total_processed += processed
            total_skipped += skipped
            projects_processed += 1
        except Exception as e:
            logger.error(f"Error migrating project {project_name}: {e}")
            projects_failed += 1
            continue

    logger.info("✓ Migration completed successfully!")
    logger.info("Summary:")
    logger.info(f"  - Total projects in data: {total_projects}")
    logger.info(f"  - Projects processed successfully: {projects_processed}")
    logger.info(f"  - Projects failed: {projects_failed}")
    logger.info(f"  - Total items loaded: {total_processed}")
    logger.info(f"  - Total items skipped: {total_skipped}")
    logger.info(f"  - Total items processed: {total_processed + total_skipped}")


if __name__ == "__main__":
    main()
