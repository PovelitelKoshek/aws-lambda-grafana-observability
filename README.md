# AWS Lambda Grafana Observability

This repository contains a proof-of-concept for sending telemetry from AWS Lambda directly to Grafana Cloud without using EC2, VM, self-hosted Grafana Alloy, or a separate OpenTelemetry Collector server.

## Project goal

The main goal of this project is to test direct telemetry delivery from AWS Lambda to Grafana Cloud.

The project consists of two parts:

1. Sending Lambda logs directly to Grafana Cloud Loki using a custom Lambda Extension.
2. Sending Lambda traces and metrics directly to Grafana Cloud using OpenTelemetry SDK and OTLP endpoint.

## General architecture

```
Logs:
AWS Lambda → Custom Lambda Extension → AWS Lambda Telemetry API → Loki HTTP Push API → Grafana Cloud Loki

Traces and Metrics:
AWS Lambda Go code → OpenTelemetry SDK → Grafana Cloud OTLP endpoint → Grafana Cloud Traces / Tempo + Metrics / Mimir
```

Repository structure

docs/```
Documentation, architecture, sources, Grafana queries and conclusions.```

logs-loki-extension/```
Custom Go Lambda Extension for sending Lambda logs directly to Grafana Cloud Loki.```

traces-metrics-otel/```
Go Lambda function with OpenTelemetry SDK for sending traces and metrics directly to Grafana Cloud OTLP endpoint.```

scripts/```
Shell scripts and CloudShell commands used for build, deploy and testing.```

## Part 1: Logs to Loki

The first part of the project is a custom AWS Lambda Extension written in Go.

The extension subscribes to AWS Lambda Telemetry API, receives function and platform logs, formats them and sends them directly to Grafana Cloud Loki through Loki HTTP Push API.

```AWS Lambda > Custom Go Lambda Extension > AWS Lambda Telemetry API > Grafana Cloud Loki```

This part is stored in:

logs-loki-extension/

## Part 2: Traces and Metrics through OTLP

The second part of the project is a Go Lambda function with OpenTelemetry SDK embedded directly into the main function code.

The function creates traces and custom metrics, then sends them directly to Grafana Cloud OTLP endpoint.

```AWS Lambda Go function > OpenTelemetry SDK  > Grafana Cloud OTLP endpoint  > Grafana Cloud Traces / Tempo  > Grafana Cloud Metrics / Mimir```

This part is stored in:

traces-metrics-otel/
Implemented telemetry
Logs

Collected through AWS Lambda Telemetry API and sent to Loki.

Traces

The Lambda function creates a trace span for handler execution:

lambda_go_demo_handler

Metrics

The Lambda function sends custom metrics:

lambda_requests_total
lambda_errors_total
lambda_cold_starts_total
lambda_duration_ms
lambda_simulated_work_ms

## Result

The proof-of-concept confirmed that AWS Lambda can send telemetry directly to Grafana Cloud.

In Grafana Cloud:

logs   → Grafana Cloud Loki
traces → grafanacloud-...-traces
metrics → grafanacloud-...-prom

The architecture does not require EC2, VM, self-hosted Alloy, or a separate OpenTelemetry Collector server.
