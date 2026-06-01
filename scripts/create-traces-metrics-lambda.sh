#!/bin/bash

set -e

FUNCTION_NAME="lambda-go-otel-grafana"
ROLE_NAME="lambda-go-otel-grafana-role"
ZIP_FILE="traces-metrics-otel/lambda-go-otel-grafana.zip"

cat > trust-policy.json <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

aws iam create-role \
  --role-name "$ROLE_NAME" \
  --assume-role-policy-document file://trust-policy.json || true

aws iam attach-role-policy \
  --role-name "$ROLE_NAME" \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

sleep 10

ROLE_ARN=$(aws iam get-role \
  --role-name "$ROLE_NAME" \
  --query 'Role.Arn' \
  --output text)

aws lambda create-function \
  --function-name "$FUNCTION_NAME" \
  --runtime provided.al2023 \
  --handler bootstrap \
  --architectures x86_64 \
  --role "$ROLE_ARN" \
  --zip-file fileb://"$ZIP_FILE" \
  --timeout 10 \
  --memory-size 512 || \
aws lambda update-function-code \
  --function-name "$FUNCTION_NAME" \
  --zip-file fileb://"$ZIP_FILE"

echo "Lambda deployment completed"
