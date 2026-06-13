package authclient

import "time"

type Identity struct {
	UserID      int64  `json:"id"`
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
}

type VerifyResult struct {
	Valid    bool
	Identity Identity
}

type Config struct {
	BaseURL       string
	InternalToken string
	Timeout       time.Duration
}
