#!/bin/bash

set -e

cd traces-metrics-otel

go mod tidy
go build -o bootstrap .
zip lambda-4-go-otel-grafana.zip bootstrap

echo "Build completed: lambda-4-go-otel-grafana.zip"
