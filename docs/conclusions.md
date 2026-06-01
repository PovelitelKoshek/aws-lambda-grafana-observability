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
Lambda Go code → OpenTelemetry SDK → OTLP endpoint → Metrics

This approach keeps the architecture serverless and avoids maintaining an additional telemetry server.
