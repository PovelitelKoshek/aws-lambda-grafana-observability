package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const listenerPort = "4243"

type TelemetryEvent struct {
	Time   string      `json:"time"`
	Type   string      `json:"type"`
	Record interface{} `json:"record"`
}

type LokiPayload struct {
	Streams []LokiStream `json:"streams"`
}

type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func main() {
	log.Println("TEST3_EXTENSION: starting")

	runtimeAPI := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	if runtimeAPI == "" {
		log.Println("TEST3_EXTENSION: AWS_LAMBDA_RUNTIME_API is empty")
		return
	}

	extensionName := filepath.Base(os.Args[0])

	extensionID, err := registerExtension(runtimeAPI, extensionName)
	if err != nil {
		log.Printf("TEST3_EXTENSION: register error: %v", err)
		return
	}

	log.Println("TEST3_EXTENSION: registered successfully")

	startTelemetryListener()

	err = subscribeTelemetry(runtimeAPI, extensionID)
	if err != nil {
		log.Printf("TEST3_EXTENSION: telemetry subscribe error: %v", err)
		return
	}

	log.Println("TEST3_EXTENSION: telemetry subscribed successfully")

	for {
		eventType, err := nextEvent(runtimeAPI, extensionID)
		if err != nil {
			log.Printf("TEST3_EXTENSION: next event error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if eventType == "INVOKE" {
			log.Println("TEST3_EXTENSION: invoke finished, waiting for telemetry delivery")
			time.Sleep(20 * time.Second)
		}

		if eventType == "SHUTDOWN" {
			log.Println("TEST3_EXTENSION: shutdown")
			time.Sleep(2 * time.Second)
			return
		}
	}
}

func registerExtension(runtimeAPI string, extensionName string) (string, error) {
	url := "http://" + runtimeAPI + "/2020-01-01/extension/register"

	body := []byte(`{"events":["INVOKE","SHUTDOWN"]}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Extension-Name", extensionName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("register failed: %s body=%s", resp.Status, string(respBody))
	}

	extensionID := resp.Header.Get("Lambda-Extension-Identifier")
	if extensionID == "" {
		return "", fmt.Errorf("missing Lambda-Extension-Identifier")
	}

	return extensionID, nil
}

func startTelemetryListener() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("TEST3_EXTENSION: read telemetry body error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var events []TelemetryEvent
		if err := json.Unmarshal(body, &events); err != nil {
			log.Printf("TEST3_EXTENSION: telemetry json error: %v body=%s", err, string(body))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("TEST3_EXTENSION: received telemetry events: %d", len(events))

		for _, event := range events {
			sendEventToLoki(event)
		}

		w.WriteHeader(http.StatusOK)
	})

	go func() {
		log.Println("TEST3_EXTENSION: listener started on port " + listenerPort)
		err := http.ListenAndServe(":"+listenerPort, nil)
		if err != nil {
			log.Printf("TEST3_EXTENSION: listener error: %v", err)
		}
	}()
}

func subscribeTelemetry(runtimeAPI string, extensionID string) error {
	url := "http://" + runtimeAPI + "/2022-07-01/telemetry"

	payload := map[string]interface{}{
		"schemaVersion": "2022-07-01",
		"types": []string{
			"platform",
			"function",
		},
		"buffering": map[string]interface{}{
			"maxItems":  1000,
			"maxBytes":  262144,
			"timeoutMs": 100,
		},
		"destination": map[string]interface{}{
			"protocol": "HTTP",
			"URI":      "http://sandbox.localdomain:" + listenerPort,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Lambda-Extension-Identifier", extensionID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telemetry subscribe failed: %s body=%s", resp.Status, string(respBody))
	}

	return nil
}

func nextEvent(runtimeAPI string, extensionID string) (string, error) {
	url := "http://" + runtimeAPI + "/2020-01-01/extension/event/next"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Lambda-Extension-Identifier", extensionID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&payload)

	eventType, _ := payload["eventType"].(string)
	return eventType, nil
}

func sendEventToLoki(event TelemetryEvent) {
	lokiURL := os.Getenv("LOKI_PUSH_URL")
	lokiUsername := os.Getenv("LOKI_USERNAME")
	lokiToken := os.Getenv("GRAFANA_CLOUD_TOKEN")

	if lokiURL == "" || lokiUsername == "" || lokiToken == "" {
		log.Println("TEST3_EXTENSION: missing Loki environment variables")
		return
	}

	timestamp := time.Now().UnixNano()

	if event.Time != "" {
		parsedTime, err := time.Parse(time.RFC3339Nano, event.Time)
		if err == nil {
			timestamp = parsedTime.UnixNano()
		}
	}

	line := stringifyRecord(event.Record)

	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	region := os.Getenv("AWS_REGION")

	if functionName == "" {
		functionName = "unknown"
	}

	if region == "" {
		region = "unknown"
	}

	payload := LokiPayload{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"job":           "aws-lambda",
					"function_name": sanitizeLabelValue(functionName),
					"event_type":    sanitizeLabelValue(event.Type),
					"region":        sanitizeLabelValue(region),
					"source":        "lambda-extension",
				},
				Values: [][]string{
					{
						strconv.FormatInt(timestamp, 10),
						line,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("TEST3_EXTENSION: loki json marshal error: %v", err)
		return
	}

	req, err := http.NewRequest("POST", lokiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("TEST3_EXTENSION: loki request error: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(lokiUsername, lokiToken)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("TEST3_EXTENSION: loki send error: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("TEST3_EXTENSION: loki bad response: %s body=%s", resp.Status, string(respBody))
		return
	}

	log.Printf("TEST3_EXTENSION: sent event to Loki, type=%s", event.Type)
}

func stringifyRecord(record interface{}) string {
	switch v := record.(type) {
	case string:
		return v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func sanitizeLabelValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	return value
}
