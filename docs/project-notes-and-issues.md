# Project Notes and Issues

## General idea

The goal of the project was to check whether AWS Lambda can send telemetry directly to Grafana Cloud without using an additional EC2 instance, virtual machine, self-hosted Grafana Alloy, or a separate OpenTelemetry Collector.

The final architecture was split into two parts:

```
Logs:
AWS Lambda → Lambda Extension → AWS Lambda Telemetry API → Loki HTTP Push API → Grafana Cloud Loki

Traces and Metrics:
AWS Lambda Go code → OpenTelemetry SDK → Grafana Cloud OTLP endpoint → Grafana Cloud Traces / Tempo + Metrics / Mimir
```

This split was chosen because logs and platform events can be collected externally through Lambda Telemetry API, while traces and custom metrics are better created inside the application code, where the function has access to execution context, errors, duration and custom attributes.

## Where Grafana Cloud configuration was taken from

For traces and metrics, Grafana Cloud provides OTLP connection information in the OpenTelemetry configuration section.

In Grafana Cloud, the needed information was taken from:

```
Grafana Cloud → Stack → OpenTelemetry → Configure OpenTelemetry
```

From there, the following values were used:

```
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_EXPORTER_OTLP_HEADERS
```

The endpoint defines where the Lambda sends OpenTelemetry data.

The headers contain authorization data for Grafana Cloud. These values were added as AWS Lambda environment variables.

The token was created in Grafana Cloud with permissions for writing telemetry data, mainly:

```
metrics:write
traces:write
logs:write
```

## Why AWS CloudShell was used

AWS CloudShell was used because it already has AWS CLI configured for the current AWS account and region.

It was used to:

```
create the Go project
install Go dependencies
build the Lambda binary
package the binary into a zip file
create IAM Role
create or update AWS Lambda
invoke Lambda for testing
```

For Go Lambda with provided.al2023, the Go source code cannot be uploaded directly as main.go. It must first be compiled into an executable binary.

The binary was named:

```
bootstrap
```

This is required for AWS Lambda custom runtime. The bootstrap file was then packed into a zip archive and uploaded to Lambda.

## Issue 1: Go dependencies and OpenTelemetry versions

During the build process, there were problems with Go modules and OpenTelemetry package versions.

At first, Go tried to pull newer OpenTelemetry versions that required a newer Go toolchain. This caused build errors and CloudShell storage issues.

The solution was to fix the OpenTelemetry versions in go.mod and use versions compatible with the available Go environment.

The final approach was:

```
go mod tidy
go build -o bootstrap .
```

After correcting dependencies, the bootstrap binary was created successfully.

## Issue 2: bootstrap file was not created

At one stage, the bootstrap file did not appear after running the build command.

The reason was that the project files were not fully correct: go.mod and later main.go had formatting or syntax issues.

The solution was to check the files, fix go.mod, rewrite main.go correctly and then rebuild the project.

After that, the command:

```
go build -o bootstrap .
```

successfully created the bootstrap file.

## Issue 3: Choosing the correct Lambda runtime

For the Go Lambda, the runtime used was:

```
provided.al2023
```

This is an OS-only runtime based on Amazon Linux 2023.

It was used because the Go code was compiled into a standalone executable file. AWS Lambda runs the bootstrap executable inside this runtime.

The connection is:

```
Go source code → compiled binary bootstrap → zip archive → Lambda with provided.al2023
```

So provided.al2023 provides the execution environment, and bootstrap is the file that Lambda runs.

## Issue 4: Finding traces in Grafana

At first, the expected datasource name Tempo was not visible in Grafana Explore.

The solution was to check the available Grafana Cloud datasources. In Grafana Cloud, the traces datasource was not named exactly Tempo, but had a name like:

```
grafanacloud-...-traces
```

This datasource is backed by Tempo, even if the visible name is different.

After selecting this datasource, traces appeared successfully.

## Issue 5: Finding metrics in Grafana

Metrics were not checked through the traces datasource.

For metrics, the correct datasource was the Grafana Cloud Prometheus-compatible datasource, with a name like:

```
grafanacloud-...-prom
```

The first test query was:

```
{__name__=~"lambda4.*"}
```

This helped find all custom metrics sent by the Lambda function.

## Issue 6: Metrics needed to be more meaningful

The first working version mainly proved that OpenTelemetry export worked.

After that, custom metrics were improved to make Grafana dashboards more useful.

The Lambda function was updated to send:

```
lambda4_requests_total
lambda4_errors_total
lambda4_cold_starts_total
lambda4_duration_ms
lambda4_simulated_work_ms
```

These metrics allow building dashboards for:

```
total requests
request rate
errors
cold starts
average duration
P95 duration
```

## Issue 7: Grafana dashboards

After traces and metrics appeared in Grafana Cloud, dashboards were created separately for metrics and traces.

Metrics dashboards:

```
Lambda 4 Metrics Overview
Lambda 4 Metrics Performance
```

Traces dashboards:

```
Lambda 4 Traces Overview
Lambda 4 Traces Errors and Slow Requests
```

This separation was useful because metrics and traces are different telemetry types and are stored in different Grafana Cloud backends.

## Final result

The project confirmed that direct telemetry delivery from AWS Lambda to Grafana Cloud works.

The final result:
```
Logs → Loki
Traces → Grafana Cloud Traces / Tempo
Metrics → Grafana Cloud Metrics / Mimir
```
