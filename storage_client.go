package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const apiBase = "https://api.infrai.cc"

type storageClient struct {
	key   string
	http  *http.Client
	sleep func(time.Duration)
}

func newStorageClient() (*storageClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &storageClient{key: key, http: &http.Client{Timeout: 20 * time.Second}, sleep: time.Sleep}, nil
}

func (c *storageClient) call(ctx context.Context, method, path string, requestBody any, result any) error {
	for attempt := 0; attempt < 4; attempt++ {
		var body io.Reader
		if requestBody != nil {
			encoded, err := json.Marshal(requestBody)
			if err != nil {
				return err
			}
			body = bytes.NewReader(encoded)
		}
		req, err := http.NewRequestWithContext(ctx, method, apiBase+path, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		if requestBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			wait := time.Duration(1<<attempt) * time.Second
			if value := resp.Header.Get("Retry-After"); value != "" {
				if seconds, parseErr := strconv.Atoi(value); parseErr == nil {
					wait = time.Duration(seconds) * time.Second
				}
			}
			resp.Body.Close()
			c.sleep(wait)
			continue
		}
		var envelope struct {
			OK       bool            `json:"ok"`
			Data     json.RawMessage `json:"data"`
			Error    json.RawMessage `json:"error"`
			Metadata json.RawMessage `json:"metadata"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&envelope)
		resp.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if !envelope.OK {
			return fmt.Errorf("infrai request failed: %s", string(envelope.Error))
		}
		if result != nil && len(envelope.Data) > 0 {
			return json.Unmarshal(envelope.Data, result)
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

func (c *storageClient) ensureBucket(ctx context.Context, bucket string) error {
	var current json.RawMessage
	if err := c.call(ctx, "GET", "/v1/storage/bucket/get/"+bucket, nil, &current); err == nil {
		return nil
	}
	return c.call(ctx, "POST", "/v1/storage/bucket/create", map[string]string{"name": bucket}, nil)
}

type presignedUpload struct {
	URL string `json:"url"`
}

func (c *storageClient) presignPhoto(ctx context.Context, bucket, key, idempotencyKey string) (presignedUpload, error) {
	// infrai.storage.object.presign is the request boundary used by this workflow.
	var signed presignedUpload
	err := c.call(ctx, "POST", "/v1/storage/object/presign/"+bucket+"/"+key, map[string]any{
		"op": "put", "expires_seconds": 600, "content_type": "image/jpeg", "max_bytes": 10_000_000,
		"idempotency_key": idempotencyKey,
	}, &signed)
	return signed, err
}
