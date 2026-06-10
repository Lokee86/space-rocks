package playerdata

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type RailsStoreConfig struct {
	BaseURL       string
	InternalToken string
	BearerToken   string
}

type RailsStore struct {
	BaseURL       string
	internalToken string
	bearerToken   string
	httpClient    *http.Client
}

func NewRailsStore(config RailsStoreConfig) (*RailsStore, error) {
	if strings.TrimSpace(config.BaseURL) == "" {
		return nil, errors.New("base_url is required")
	}

	return &RailsStore{
		BaseURL:       strings.TrimRight(config.BaseURL, "/"),
		internalToken: config.InternalToken,
		bearerToken:   config.BearerToken,
		httpClient:    http.DefaultClient,
	}, nil
}

func (s *RailsStore) client() *http.Client {
	if s.httpClient != nil {
		return s.httpClient
	}
	return http.DefaultClient
}

func (s *RailsStore) newJSONRequest(method, path string, body any) (*http.Request, error) {
	requestURL, err := url.JoinPath(s.BaseURL, path)
	if err != nil {
		return nil, err
	}

	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(payload)
	}

	request, err := http.NewRequest(method, requestURL, requestBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	if strings.HasPrefix(path, "/internal") {
		if s.internalToken != "" {
			request.Header.Set("Authorization", "Bearer "+s.internalToken)
		}
		return request, nil
	}

	if s.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+s.bearerToken)
	}

	return request, nil
}

func (s *RailsStore) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	if identity.IdentityKind != IdentityKindAuthenticatedAccount {
		return protocol.PlayerDataStats{}, false, errors.New("identity_kind must be authenticated_account")
	}
	if identity.AccountID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("account_id is required")
	}
	if s.bearerToken == "" {
		return protocol.PlayerDataStats{}, false, errors.New("bearer token is required")
	}

	request, err := s.newJSONRequest(http.MethodGet, "/api/player/stats", nil)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	response, err := s.client().Do(request)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return protocol.PlayerDataStats{}, false, errors.New("unexpected status")
	}

	var decoded struct {
		Stats protocol.PlayerDataStats `json:"stats"`
	}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	return decoded.Stats, true, nil
}

func (s *RailsStore) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	if command.Identity.IdentityKind != IdentityKindAuthenticatedAccount {
		return protocol.PlayerDataStats{}, false, errors.New("identity_kind must be authenticated_account")
	}
	if command.Identity.AccountID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("account_id is required")
	}
	if command.ResultID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("result_id is required")
	}
	if command.MatchID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("match_id is required")
	}
	if s.internalToken == "" {
		return protocol.PlayerDataStats{}, false, errors.New("internal token is required")
	}

	request, err := s.newJSONRequest(http.MethodPost, "/internal/player-data/match-results", struct {
		ResultID   string `json:"result_id"`
		MatchID    string `json:"match_id"`
		AccountID  string `json:"account_id"`
		Score      int    `json:"score"`
		ShipDeaths int    `json:"ship_deaths"`
		Won        bool   `json:"won"`
	}{
		ResultID:   command.ResultID,
		MatchID:    command.MatchID,
		AccountID:  command.Identity.AccountID,
		Score:      command.Score,
		ShipDeaths: command.ShipDeaths,
		Won:        command.Won,
	})
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	response, err := s.client().Do(request)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return protocol.PlayerDataStats{}, false, errors.New("unexpected status")
	}

	var decoded struct {
		Accepted  bool                     `json:"accepted"`
		Duplicate bool                     `json:"duplicate"`
		Stats     protocol.PlayerDataStats `json:"stats"`
		Error     string                   `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	if !decoded.Accepted {
		if decoded.Error != "" {
			return protocol.PlayerDataStats{}, false, errors.New(decoded.Error)
		}
		return protocol.PlayerDataStats{}, false, errors.New("record match result rejected")
	}

	return decoded.Stats, decoded.Duplicate, nil
}
