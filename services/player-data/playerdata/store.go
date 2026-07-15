package playerdata

import (
	"errors"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type FailureClass string

const (
	FailureClassUpstreamUnavailable FailureClass = "upstream_unavailable"
	FailureClassUnexpectedStatus    FailureClass = "unexpected_status"
	FailureClassDecodeFailed        FailureClass = "decode_failed"
	FailureClassAuthentication      FailureClass = "authentication_failed"
	FailureClassTransaction         FailureClass = "transaction_failed"
	FailureClassStoreUnavailable    FailureClass = "store_unavailable"
	FailureClassInvalidResponse     FailureClass = "invalid_response"
)

// ClassifiedFailure carries a stable, bounded classification while preserving
// the underlying error for local control flow. Callers must not serialize the
// underlying error into observability fields.
type ClassifiedFailure struct {
	Class FailureClass
	Err   error
}

func (e *ClassifiedFailure) Error() string {
	if e == nil {
		return ""
	}
	return string(e.Class)
}

func (e *ClassifiedFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewClassifiedFailure(class FailureClass, err error) error {
	if err == nil {
		return &ClassifiedFailure{Class: class}
	}
	return &ClassifiedFailure{Class: class, Err: err}
}

func FailureClassOf(err error) FailureClass {
	var classified *ClassifiedFailure
	if errors.As(err, &classified) {
		return classified.Class
	}
	return ""
}

type Store interface {
	LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error)
	RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error)
}

type LocalProfileStore interface {
	ListLocalProfiles() ([]LocalProfileSummary, error)
	CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (LocalProfileSummary, error)
	DeleteLocalProfile(localProfileID string) error
	UpdateLocalProfileDisplayName(localProfileID string, displayName string) (LocalProfileSummary, error)
	GetDefaultLocalProfile() (LocalProfileDefault, error)
	SetDefaultLocalProfile(identityKind string, localProfileID string) (LocalProfileDefault, error)
}

var ErrLocalProfileNotFound = errors.New("local profile not found")

var ErrLocalProfileUnavailable = errors.New("local profile management is unavailable")

type LocalProfileSummary struct {
	LocalProfileID string
	DisplayName    string
}

type LocalProfileDefault struct {
	IdentityKind   string
	LocalProfileID string
	DisplayName    string
}
