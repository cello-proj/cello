#!/usr/bin/env -S uv run
# /// script
# requires-python = ">=3.12"
# dependencies = ["boto3"]
# ///

import json
import boto3
import argparse
from boto3.dynamodb.types import TypeDeserializer


def from_dynamo_json(item):
    """Convert DynamoDB typed JSON (AttributeValue format) to plain Python dict."""
    deserializer = TypeDeserializer()
    return {k: deserializer.deserialize(v) for k, v in item.items()}


def main():
    parser = argparse.ArgumentParser(
        description="Dump all items from a DynamoDB table to a JSON file."
    )
    parser.add_argument("--table", required=True, help="DynamoDB table name")
    parser.add_argument("--output", required=True, help="Output JSON file path")
    parser.add_argument("--region", required=True, help="AWS region name")

    args = parser.parse_args()

    dynamodb = boto3.client("dynamodb", region_name=args.region)
    paginator = dynamodb.get_paginator("scan")

    with open(args.output, "w") as f:
        f.write("[\n")
        first = True

        for page in paginator.paginate(TableName=args.table):
            for item in page["Items"]:
                if not first:
                    f.write(",\n")
                else:
                    first = False
                json.dump(from_dynamo_json(item), f)

        f.write("\n]\n")

    print(f"Dumped table '{args.table}' from region '{args.region}' to {args.output}")


if __name__ == "__main__":
    main()
