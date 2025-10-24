#!/usr/bin/env -S uv run
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "boto3",
# ]
# ///

"""
DynamoDB Timestamp Migration Script

Migrates timestamp formats from PostgreSQL-style (ending in +00) 
to ISO 8601 format (ending in Z) in DynamoDB records.
"""

import argparse
import json
import sys
from datetime import datetime
from typing import Any

import boto3
from botocore.exceptions import ClientError


def parse_arguments() -> argparse.Namespace:
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(
        description="Migrate DynamoDB timestamps from PostgreSQL format to ISO 8601 format"
    )
    parser.add_argument(
        "--file",
        required=True,
        help="Path to JSON input file containing records",
    )
    parser.add_argument(
        "--table",
        required=True,
        help="DynamoDB table name",
    )
    parser.add_argument(
        "--region",
        required=True,
        help="AWS region",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be updated without making actual changes",
    )
    return parser.parse_args()


def should_skip_record(record: dict[str, Any]) -> tuple[bool, str | None]:
    """
    Check if a record should be skipped.
    
    Returns:
        Tuple of (should_skip, reason)
    """
    # Skip METADATA records
    if record.get("sk") == "METADATA":
        return True, "METADATA record"
    
    # Check if required fields exist
    if "created_at" not in record or "expires_at" not in record:
        return True, "Missing timestamp fields"
    
    created_at = record["created_at"]
    expires_at = record["expires_at"]
    
    # Skip records with timestamps already in ISO format (ending with Z)
    if created_at.endswith("Z") or expires_at.endswith("Z"):
        return True, "Already in ISO format (ends with Z)"
    
    # Only process records with timestamps ending in +00
    if not (created_at.endswith("+00") and expires_at.endswith("+00")):
        return True, "Timestamps don't end with +00"
    
    return False, None


def convert_timestamp(timestamp: str) -> str:
    """
    Convert PostgreSQL timestamp to ISO 8601/RFC3339 format with nanosecond precision.
    
    Input: "2022-02-02 18:01:49.345261+00"
    Output: "2022-02-02T18:01:49.345261000Z"
    """
    dt = datetime.fromisoformat(timestamp)
    # Python datetime has microseconds (6 digits), pad to nanoseconds (9 digits) and add Z
    return dt.strftime("%Y-%m-%dT%H:%M:%S") + f".{dt.microsecond:06d}000Z"


def update_dynamodb_record(
    dynamodb_client: Any,
    table_name: str,
    pk: str,
    sk: str,
    created_at: str,
    expires_at: str,
    dry_run: bool,
) -> tuple[bool, str | None]:
    """
    Update a DynamoDB record with new timestamp values.
    
    Returns:
        Tuple of (success, error_message)
    """
    if dry_run:
        return True, None
    
    try:
        dynamodb_client.update_item(
            TableName=table_name,
            Key={
                "pk": {"S": pk},
                "sk": {"S": sk},
            },
            UpdateExpression="SET created_at = :created_at, expires_at = :expires_at",
            ExpressionAttributeValues={
                ":created_at": {"S": created_at},
                ":expires_at": {"S": expires_at},
            },
            ConditionExpression="attribute_exists(pk) AND attribute_exists(sk)",
        )
        return True, None
    except ClientError as e:
        error_code = e.response["Error"]["Code"]
        error_message = e.response["Error"]["Message"]
        return False, f"{error_code}: {error_message}"


def main() -> None:
    """Main execution function."""
    args = parse_arguments()
    
    # Print configuration
    print("Configuration:")
    print(f"  Input file: {args.file}")
    print(f"  DynamoDB table: {args.table}")
    print(f"  Region: {args.region}")
    print(f"  Dry run: {args.dry_run}")
    print()
    
    # Load records from file
    print(f"Loading records from {args.file}...")
    try:
        with open(args.file, "r") as f:
            records = json.load(f)
    except FileNotFoundError:
        print(f"Error: File not found: {args.file}", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON in file: {e}", file=sys.stderr)
        sys.exit(1)
    
    print(f"Loaded {len(records)} total records")
    print()
    
    # Initialize DynamoDB client
    dynamodb_client = boto3.client("dynamodb", region_name=args.region)
    
    # Statistics
    total_records = len(records)
    skipped_count = 0
    processed_count = 0
    success_count = 0
    failure_count = 0
    
    # Process each record
    print("Processing records:")
    print("-" * 80)
    
    for i, record in enumerate(records, 1):
        # Check if record should be skipped
        should_skip, skip_reason = should_skip_record(record)
        
        if should_skip:
            skipped_count += 1
            continue
        
        # Extract fields
        pk = record["pk"]
        sk = record["sk"]
        old_created_at = record["created_at"]
        old_expires_at = record["expires_at"]
        
        # Convert timestamps
        new_created_at = convert_timestamp(old_created_at)
        new_expires_at = convert_timestamp(old_expires_at)
        
        # Print conversion
        print(f"[{pk}][{sk}]:")
        print(f"  created_at: {old_created_at} → {new_created_at}")
        print(f"  expires_at: {old_expires_at} → {new_expires_at}")
        
        # Update DynamoDB
        success, error = update_dynamodb_record(
            dynamodb_client,
            args.table,
            pk,
            sk,
            new_created_at,
            new_expires_at,
            args.dry_run,
        )
        
        processed_count += 1
        
        if success:
            success_count += 1
            status = "DRY RUN - Would update" if args.dry_run else "✓ Updated"
            print(f"  {status}")
        else:
            failure_count += 1
            print(f"  ✗ Failed: {error}")
        
        print()
    
    # Print summary
    print("-" * 80)
    print("Summary:")
    print(f"  Total records: {total_records}")
    print(f"  Skipped: {skipped_count}")
    print(f"  Processed: {processed_count}")
    print(f"  Successful: {success_count}")
    print(f"  Failed: {failure_count}")
    
    if args.dry_run:
        print()
        print("DRY RUN MODE - No actual updates were made to DynamoDB")
    
    # Exit with error code if there were failures
    if failure_count > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()

