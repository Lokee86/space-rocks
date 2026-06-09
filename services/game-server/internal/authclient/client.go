package authclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

type verifyTokenRequest struct {
	Token string `json:"token"`
}

type verifyTokenResponse struct {
	Valid bool     `json:"valid"`
	User  Identity `json:"user"`
}

func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.BaseURL) == "" {
		return nil, errors.New("authclient: base url is required")
	}
	if strings.TrimSpace(config.InternalToken) == "" {
		return nil, errors.New("authclient: internal token is required")
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	return &Client{
		baseURL:       strings.TrimRight(config.BaseURL, "/"),
		internalToken: config.InternalToken,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) VerifyToken(ctx context.Context, rawToken string) (VerifyResult, error) {
	reqBody, err := json.Marshal(verifyTokenRequest{Token: rawToken})
	if err != nil {
		return VerifyResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/auth/verify-token", strings.NewReader(string(reqBody)))
	if err != nil {
		return VerifyResult{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.internalToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return VerifyResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return VerifyResult{}, fmt.Errorf("authclient: verify token request failed: %s", resp.Status)
	}

	var decoded verifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return VerifyResult{}, err
	}
	if !decoded.Valid {
		return VerifyResult{Valid: false}, nil
	}

	return VerifyResult{
		Valid:    true,
		Identity: decoded.User,
	}, nil
}
