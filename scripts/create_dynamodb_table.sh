#!/bin/bash

# Create table with explicit parameters
aws dynamodb create-table \
  --endpoint-url http://localhost:8000 \
  --region us-west-2 \
  --table-name cello \
  --attribute-definitions \
  AttributeName=pk,AttributeType=S \
  AttributeName=sk,AttributeType=S \
  --key-schema \
  AttributeName=pk,KeyType=HASH \
  AttributeName=sk,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST
