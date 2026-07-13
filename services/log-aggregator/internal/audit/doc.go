// Package audit owns promotion of explicitly required events into immutable audit records.
// It validates the source projection, sanitizes a defensive payload copy, and
// persists the resulting versioned audit record through the Store seam.
package audit
