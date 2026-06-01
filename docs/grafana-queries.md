# Grafana Queries

This file contains PromQL and TraceQL queries used in Grafana Cloud.

## Metrics datasource

Example datasource name:

```
grafanacloud-...-prom
```


## All project metrics:

Total invocations:

```
sum(lambda_requests_total)
```


Invocation rate:
```
sum(rate(lambda_requests_total[5m]))
```

Total errors:
```
sum(lambda_errors_total)
```

Error rate:
```
sum(rate(lambda_errors_total[5m]))
```

Cold starts:
```
sum(lambda_cold_starts_total)
```

Average Lambda duration:
```
sum(rate(lambda_duration_ms_sum[5m])) / sum(rate(lambda_duration_ms_count[5m]))
```

P95 Lambda duration:
```
histogram_quantile(0.95, sum(rate(lambda_duration_ms_bucket[5m])) by (le))
```

Average simulated work duration:
```
sum(rate(lambda_simulated_work_ms_sum[5m])) / sum(rate(lambda_simulated_work_ms_count[5m]))
```

## Traces datasource

Example datasource name:
```
grafanacloud-...-traces
```

All traces:
```
{ resource.service.name = "lambda-go-otel-grafana" }
```

Handler spans:
```
{ resource.service.name = "lambda-go-otel-grafana" && name = "lambda_go_demo_handler" }
```

Slow traces:
```
{ resource.service.name = "lambda-go-otel-grafana" && duration > 500ms }
```

Very slow traces:
```
{ resource.service.name = "lambda-go-otel-grafana" && duration > 1s }
```

Error traces:
```
{ resource.service.name = "lambda-go-otel-grafana" && status = error }
```
