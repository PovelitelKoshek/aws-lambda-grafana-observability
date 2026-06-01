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
