# Architecture

## Main idea

The project checks whether AWS Lambda can send telemetry directly to Grafana Cloud without using EC2, VM, self-hosted Grafana Alloy, or a separate OpenTelemetry Collector server.

The project uses two different approaches depending on the telemetry type:

- logs are collected by a Lambda Extension
- traces and metrics are created inside the Go Lambda code using OpenTelemetry SDK

## Final architecture

```text
AWS Lambda
   ├── Logs → Lambda Extension → Loki HTTP Push API → Grafana Cloud Loki
   └── Traces/Metrics → OpenTelemetry SDK → Grafana Cloud OTLP endpoint → Tempo + Metrics/Mimir
Logs flow

For logs, a custom Go Lambda Extension is used.

The extension works next to the main Lambda function. It subscribes to AWS Lambda Telemetry API and receives platform and function logs.

Then it formats these logs and sends them directly to Grafana Cloud Loki through Loki HTTP Push API.

AWS Lambda
    ↓
Custom Go Lambda Extension
    ↓
AWS Lambda Telemetry API
    ↓
Loki HTTP Push API
    ↓
Grafana Cloud Loki

This approach is suitable for logs because logs and platform events can be collected outside the main function code.

Traces and metrics flow

For traces and metrics, OpenTelemetry SDK is embedded directly into the Go Lambda function.

AWS Lambda Go function
    ↓
OpenTelemetry SDK
    ↓
Grafana Cloud OTLP endpoint
    ↓
Grafana Cloud Traces / Tempo
    ↓
Grafana Cloud Metrics / Mimir

This approach is better for traces and custom metrics because the main application code knows what happens inside the function: request start, errors, duration, cold start and custom attributes.

Why two approaches are used

Logs can be collected from outside the main function using Lambda Telemetry API, so a Lambda Extension is suitable for logs.

Traces and custom metrics require application-level context, so OpenTelemetry SDK is added directly into the main Go Lambda code.

Used Grafana Cloud backends
Loki   → logs
Tempo  → traces
Mimir / Prometheus-compatible backend → metrics
Result

The project confirms that AWS Lambda can send logs, traces and metrics directly to Grafana Cloud using a serverless architecture.


Commit changes.

---

# Шаг 4. Создай `docs/cloudshell-commands.md`

Имя файла:

```text
docs/cloudshell-commands.md

Вставь:

# AWS CloudShell Commands

This file contains the main AWS CloudShell commands used during the project.

## Build Go Lambda for traces and metrics

```bash
cd ~/lambda-4-go-otel-grafana

go mod tidy

go build -o bootstrap .

zip lambda-4-go-otel-grafana.zip bootstrap
Create IAM trust policy
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
Create IAM role
aws iam create-role \
  --role-name lambda-4-go-otel-grafana-role \
  --assume-role-policy-document file://trust-policy.json
Attach Lambda basic execution policy
aws iam attach-role-policy \
  --role-name lambda-4-go-otel-grafana-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
Get role ARN
ROLE_ARN=$(aws iam get-role \
  --role-name lambda-4-go-otel-grafana-role \
  --query 'Role.Arn' \
  --output text)

echo $ROLE_ARN
Create Lambda function
aws lambda create-function \
  --function-name lambda-4-go-otel-grafana \
  --runtime provided.al2023 \
  --handler bootstrap \
  --architectures x86_64 \
  --role "$ROLE_ARN" \
  --zip-file fileb://lambda-4-go-otel-grafana.zip \
  --timeout 10 \
  --memory-size 512
Configure environment variables

Grafana Cloud provides:

OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_EXPORTER_OTLP_HEADERS

Example environment configuration:

cat > env.json <<EOF
{
  "Variables": {
    "OTEL_SERVICE_NAME": "lambda-4-go-otel-grafana",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "$OTEL_EXPORTER_OTLP_ENDPOINT",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
    "OTEL_EXPORTER_OTLP_HEADERS": "$OTEL_EXPORTER_OTLP_HEADERS",
    "OTEL_TRACES_SAMPLER": "always_on",
    "OTEL_RESOURCE_ATTRIBUTES": "deployment.environment=poc,service.namespace=aws-lambda-grafana,service.version=1.0"
  }
}
EOF

Apply environment variables:

aws lambda update-function-configuration \
  --function-name lambda-4-go-otel-grafana \
  --environment file://env.json
Invoke Lambda

Create test event:

cat > event.json <<'EOF'
{
  "hello": "grafana",
  "test": "lambda4-go"
}
EOF

Invoke Lambda:

aws lambda invoke \
  --function-name lambda-4-go-otel-grafana \
  --payload fileb://event.json \
  response.json

Run multiple invocations:

for i in {1..10}; do
  aws lambda invoke \
    --function-name lambda-4-go-otel-grafana \
    --payload fileb://event.json \
    response-$i.json
done

Create error test event:

cat > fail-event.json <<'EOF'
{
  "fail": true
}
EOF

Invoke error test:

aws lambda invoke \
  --function-name lambda-4-go-otel-grafana \
  --payload fileb://fail-event.json \
  fail-response.json
Notes

The following files should not be committed to GitHub:

env.json
bootstrap
*.zip
response.json
fail-response.json

Commit changes.

---

# Шаг 5. Создай `docs/grafana-queries.md`

Имя файла:

```text
docs/grafana-queries.md

Вставь:

# Grafana Queries

This file contains PromQL and TraceQL queries used in Grafana Cloud.

## Metrics datasource

Example datasource name:

```text
grafanacloud-...-prom

Find all project metrics:

{__name__=~"lambda4.*"}

Total invocations:

sum(lambda4_requests_total)

Invocation rate:

sum(rate(lambda4_requests_total[5m]))

Total errors:

sum(lambda4_errors_total)

Error rate:

sum(rate(lambda4_errors_total[5m]))

Cold starts:

sum(lambda4_cold_starts_total)

Average Lambda duration:

sum(rate(lambda4_duration_ms_sum[5m])) / sum(rate(lambda4_duration_ms_count[5m]))

P95 Lambda duration:

histogram_quantile(0.95, sum(rate(lambda4_duration_ms_bucket[5m])) by (le))

Average simulated work duration:

sum(rate(lambda4_simulated_work_ms_sum[5m])) / sum(rate(lambda4_simulated_work_ms_count[5m]))
Traces datasource

Example datasource name:

grafanacloud-...-traces

All traces:

{ resource.service.name = "lambda-4-go-otel-grafana" }

Handler spans:

{ resource.service.name = "lambda-4-go-otel-grafana" && name = "lambda4_go_demo_handler" }

Slow traces:

{ resource.service.name = "lambda-4-go-otel-grafana" && duration > 500ms }

Very slow traces:

{ resource.service.name = "lambda-4-go-otel-grafana" && duration > 1s }

Error traces:

{ resource.service.name = "lambda-4-go-otel-grafana" && status = error }
Created dashboards
Metrics dashboards
Lambda 4 Metrics Overview
Lambda 4 Metrics Performance
Traces dashboards
Lambda 4 Traces Overview
Lambda 4 Traces Errors and Slow Requests

Commit changes.

---

# Шаг 6. Создай `docs/conclusions.md`

Имя файла:

```text
docs/conclusions.md

Вставь:

# Conclusions

## Main result

The project confirmed that AWS Lambda telemetry can be sent directly to Grafana Cloud without EC2, VM, self-hosted Grafana Alloy, or a separate OpenTelemetry Collector server.

## Logs

Logs were sent using a custom Go Lambda Extension.

The extension subscribes to AWS Lambda Telemetry API and sends logs directly to Grafana Cloud Loki using Loki HTTP Push API.

## Traces and metrics

Traces and metrics were sent using OpenTelemetry SDK embedded directly into the Go Lambda function.

The function sends telemetry directly to Grafana Cloud OTLP endpoint.

## Grafana result

Traces appeared in:

```text
grafanacloud-...-traces

Metrics appeared in:

grafanacloud-...-prom

Logs appeared in Grafana Cloud Loki.

Final architecture
Logs:
Lambda → Lambda Extension → Loki

Traces:
Lambda Go code → OpenTelemetry SDK → OTLP endpoint → Tempo

Metrics:
Lambda Go code → OpenTelemetry SDK → OTLP endpoint → Metrics/Mimir
Why this approach is useful

This approach keeps the architecture serverless and avoids maintaining an additional telemetry server.

It also separates telemetry by type:

logs   → Loki
traces → Tempo
metrics → Mimir / Prometheus-compatible backend
Production note

For a real production project, the OpenTelemetry initialization should be moved into a separate package or shared module. This would make the main Lambda business logic cleaner.


Commit changes.

---

# Шаг 7. Создай `docs/sources.md`

Имя файла:

```text
docs/sources.md

Вставь:

# Sources

This file contains the main documentation sources used during the project.

## Grafana Cloud OTLP

Grafana Cloud documentation about sending OpenTelemetry data through OTLP endpoint:

https://grafana.com/docs/grafana-cloud/send-data/otlp/

https://grafana.com/docs/grafana-cloud/send-data/otlp/send-data-otlp/

## OpenTelemetry Go

OpenTelemetry Go documentation:

https://opentelemetry.io/docs/languages/go/

https://opentelemetry.io/docs/languages/go/getting-started/

OpenTelemetry Go GitHub repository:

https://github.com/open-telemetry/opentelemetry-go

## AWS Lambda Go

AWS Lambda Go handler documentation:

https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html

## AWS Lambda custom runtime

AWS Lambda custom runtime documentation:

https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html

AWS Lambda runtime walkthrough with bootstrap:

https://docs.aws.amazon.com/lambda/latest/dg/runtimes-walkthrough.html

Amazon Linux 2023 Lambda runtime:

https://aws.amazon.com/blogs/compute/introducing-the-amazon-linux-2023-runtime-for-aws-lambda/

## AWS Lambda Telemetry API

AWS Lambda Telemetry API documentation:

https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html

## Grafana Loki

Loki HTTP API documentation:

https://grafana.com/docs/loki/latest/reference/loki-http-api/

Loki push API:

https://grafana.com/docs/loki/latest/reference/loki-http-api/#push-log-entries-to-loki
