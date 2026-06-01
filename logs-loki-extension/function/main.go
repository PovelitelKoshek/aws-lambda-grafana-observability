package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event map[string]interface{}

func handler(ctx context.Context, event Event) (map[string]interface{}, error) {
	log.Println("TEST: Lambda function started")
	log.Printf("TEST: incoming event = %+v", event)
	log.Println("TEST: this log should go to Grafana Cloud Loki through Lambda Extension")

	return map[string]interface{}{
		"status":  "ok",
		"message": "Hello from Test Lambda",
		"time":    time.Now().Format(time.RFC3339),
	}, nil
}

func main() {
	lambda.Start(handler)
}
