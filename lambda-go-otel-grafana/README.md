# Traces and Metrics Lambda

This directory contains the Go Lambda function that sends traces and metrics directly to Grafana Cloud through OpenTelemetry OTLP endpoint.

## Data flow

```text
AWS Lambda Go code
    ↓
OpenTelemetry SDK
    ↓
Grafana Cloud OTLP endpoint
    ↓
Grafana Cloud Traces / Tempo
    ↓
Grafana Cloud Metrics / Mimir
Main idea

OpenTelemetry SDK is initialized directly inside the Lambda code.

The Lambda creates:

trace span for handler execution
request counter
error counter
cold start counter
duration histogram
simulated work duration histogram
Metrics
lambda_requests_total
lambda_errors_total
lambda_cold_starts_total
lambda4_duration_ms
lambda_simulated_work_ms

Trace span
lambda4_go_demo_handler

Build
go mod tidy
go build -o bootstrap .
zip lambda-4-go-otel-grafana.zip bootstrap

Environment variables
OTEL_SERVICE_NAME
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_EXPORTER_OTLP_PROTOCOL
OTEL_EXPORTER_OTLP_HEADERS
OTEL_TRACES_SAMPLER
OTEL_RESOURCE_ATTRIBUTES
