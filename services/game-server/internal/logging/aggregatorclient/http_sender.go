package aggregatorclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// batchSender is the narrow delivery seam used by the future queue worker.
type batchSender interface {
	send(context.Context, []byte) error
}

type httpBatchSender struct {
	endpointURL    string
	bearerToken    string
	requestTimeout time.Duration
	client         *http.Client
}

func newHTTPBatchSender(config Config) *httpBatchSender {
	return &httpBatchSender{
		endpointURL:    config.EndpointURL,
		bearerToken:    config.BearerToken,
		requestTimeout: config.RequestTimeout,
		client:         &http.Client{Timeout: config.RequestTimeout},
	}
}

func (s *httpBatchSender) send(parent context.Context, batch []byte) error {
	ctx, cancel := context.WithTimeout(parent, s.requestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpointURL, bytes.NewReader(batch))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if s.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+s.bearerToken)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
	if readErr != nil {
		return fmt.Errorf("aggregator returned HTTP %d (reading response: %w)", response.StatusCode, readErr)
	}
	return fmt.Errorf("aggregator returned HTTP %d: %s", response.StatusCode, bytes.TrimSpace(body))
}
