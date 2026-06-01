# AWS CloudShell Commands

This file contains the main AWS CloudShell commands used during the project.

## Build Go Lambda for traces and metrics

```
bash
cd ~/lambda-go-otel-grafana

go mod tidy

go build -o bootstrap .

zip lambda-go-otel-grafana.zip bootstrap
```

## Create IAM trust policy

```
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
```

## Create IAM role
```
aws iam create-role \
  --role-name lambda-go-otel-grafana-role \
  --assume-role-policy-document file://trust-policy.json
Attach Lambda basic execution policy
aws iam attach-role-policy \
  --role-name lambda-go-otel-grafana-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
Get role ARN
ROLE_ARN=$(aws iam get-role \
  --role-name lambda-go-otel-grafana-role \
  --query 'Role.Arn' \
  --output text)

echo $ROLE_ARN
```

## Create Lambda function
``` aws lambda create-function \
  --function-name lambda-go-otel-grafana \
  --runtime provided.al2023 \
  --handler bootstrap \
  --architectures x86_64 \
  --role "$ROLE_ARN" \
  --zip-file fileb://lambda-go-otel-grafana.zip \
  --timeout 10 \
  --memory-size 512
Configure environment variables
```

## Grafana Cloud provides:

```
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_EXPORTER_OTLP_HEADERS
```

## Example environment configuration:

```
cat > env.json <<EOF
{
  "Variables": {
    "OTEL_SERVICE_NAME": "lambda-go-otel-grafana",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "$OTEL_EXPORTER_OTLP_ENDPOINT",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
    "OTEL_EXPORTER_OTLP_HEADERS": "$OTEL_EXPORTER_OTLP_HEADERS",
    "OTEL_TRACES_SAMPLER": "always_on",
    "OTEL_RESOURCE_ATTRIBUTES": "deployment.environment=poc,service.namespace=aws-lambda-grafana,service.version=1.0"
  }
}
EOF
```

## Apply environment variables:

```
aws lambda update-function-configuration \
  --function-name lambda-go-otel-grafana \
  --environment file://env.json
```

## Create test event:
```
cat > event.json <<'EOF'
{
  "hello": "grafana",
  "test": "lambda-go"
}
EOF
```

## Invoke Lambda:

```
aws lambda invoke \
  --function-name lambda-go-otel-grafana \
  --payload fileb://event.json \
  response.json
```

## Run multiple invocations:

```
for i in {1..10}; do
  aws lambda invoke \
    --function-name lambda-go-otel-grafana \
    --payload fileb://event.json \
    response-$i.json
done
```

## Create error test event:

```
cat > fail-event.json <<'EOF'
{
  "fail": true
}
EOF
```

## Invoke error test:
```
aws lambda invoke \
  --function-name lambda-go-otel-grafana \
  --payload fileb://fail-event.json \
  fail-response.json
Notes
```
