package auditprojection

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidAuditEventID = errors.New("audit: invalid audit event id")
var ErrAuditRecordNotFound = errors.New("audit: record not found")

type RecordNotFoundError struct {
	AuditEventID string
}

func (e *RecordNotFoundError) Error() string {
	return fmt.Sprintf("audit: record %q not found", e.AuditEventID)
}

func (e *RecordNotFoundError) Unwrap() error { return ErrAuditRecordNotFound }

type Store interface {
	Save(context.Context, Record) error
	Get(context.Context, string) (Record, error)
}
