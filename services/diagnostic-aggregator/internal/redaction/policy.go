package redaction

import "regexp"

// Rule describes one policy decision. Traversal code depends only on this
// seam, so contract-generated policy data can replace the defaults later.
type Rule struct {
	ID           string
	Keys         []string
	ValuePattern *regexp.Regexp
	ReasonCode   string
}

type Policy struct {
	Rules            []Rule
	Separators       []rune
	CaseInsensitive  bool
	RemoveSeparators bool
}

type Finding struct {
	RuleID              string `json:"rule_id"`
	NormalizedFieldPath string `json:"normalized_field_path"`
	ReasonCode          string `json:"reason_code"`
}

// DefaultPolicy is the current mandatory reject-event policy from the shared
// observability redaction contract.
func DefaultPolicy() Policy {
	keyRules := []struct {
		id     string
		keys   []string
		reason string
	}{
		{"forbidden_auth_transport_fields", []string{"authorization", "auth", "auth_header", "authentication", "authentication_header", "cookie", "cookies", "set_cookie"}, "forbidden_auth_data"},
		{"forbidden_credential_fields", []string{"access_token", "access_token_value", "refresh_token", "refresh_token_value", "bearer_token", "id_token", "identity_token", "oauth_code", "oauth_token", "api_key", "apikey", "api_secret", "password", "passphrase", "secret", "secret_key", "private_key", "signing_key", "encryption_key", "otp", "otp_code", "mfa_code", "one_time_code"}, "forbidden_credential_data"},
		{"forbidden_financial_fields", []string{"card_number", "cardholder_name", "card_expiry", "expiration_date", "cvv", "cvc", "pin", "bank_account", "bank_account_number", "account_number", "routing_number", "iban", "swift_code"}, "forbidden_financial_data"},
		{"forbidden_raw_content_fields", []string{"private_profile", "raw_profile", "raw_packet", "raw_packets", "packet_payload", "raw_payload", "payload", "request_body", "response_body", "body", "provider_payload", "webhook_payload", "headers", "request_headers", "response_headers"}, "forbidden_raw_content"},
	}
	policy := Policy{Separators: []rune{'_', '-', '.', '/', '\\', ' '}, CaseInsensitive: true, RemoveSeparators: true}
	for _, item := range keyRules {
		policy.Rules = append(policy.Rules, Rule{ID: item.id, Keys: item.keys, ReasonCode: item.reason})
	}
	policy.Rules = append(policy.Rules,
		Rule{ID: "bearer_credential_value", ValuePattern: regexp.MustCompile(`(?i)^\s*Bearer\s+[^\s]+\s*$`), ReasonCode: "forbidden_bearer_credential"},
		Rule{ID: "private_key_pem_header", ValuePattern: regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH |DSA |ENCRYPTED )?PRIVATE KEY-----`), ReasonCode: "forbidden_private_key_material"},
	)
	return policy
}
