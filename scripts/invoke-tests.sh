#!/bin/bash

set -e

FUNCTION_NAME="lambda-4-go-otel-grafana"

cat > event.json <<'EOF'
{
  "hello": "grafana",
  "test": "lambda4-go"
}
EOF

for i in {1..10}; do
  aws lambda invoke \
    --function-name "$FUNCTION_NAME" \
    --payload fileb://event.json \
    response-$i.json
done

cat > fail-event.json <<'EOF'
{
  "fail": true
}
EOF

aws lambda invoke \
  --function-name "$FUNCTION_NAME" \
  --payload fileb://fail-event.json \
  fail-response.json || true

echo "Test invocations completed"
