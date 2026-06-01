# Logs to Loki Lambda Extension

This directory contains the custom Go Lambda Extension used to send AWS Lambda logs directly to Grafana Cloud Loki.

## Data flow

```
AWS Lambda
    ↓
Custom Lambda Extension
    ↓
AWS Lambda Telemetry API
    ↓
Loki HTTP Push API
    ↓
Grafana Cloud Loki
```

## Main idea

The extension subscribes to AWS Lambda Telemetry API and receives function/platform logs.

Then it formats logs into Loki push format and sends them directly to Grafana Cloud Loki.

Why extension is used for logs

Logs and platform events can be collected outside the main Lambda function code. Therefore, a Lambda Extension is suitable for this part of telemetry.

Result

The extension allows sending Lambda logs to Grafana Cloud Loki without using CloudWatch as the main telemetry delivery channel.
