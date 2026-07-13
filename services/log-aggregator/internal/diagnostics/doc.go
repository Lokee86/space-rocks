// Package diagnostics owns construction and export of diagnostic bundles from aggregated events.
//
// Builder is the consumer-owned seam between diagnostic policy and storage. It
// accepts the storage query shape directly, requires a payload sanitizer before
// exposing events, and produces a versioned JSON-ready bundle without coupling
// diagnostics to HTTP query handlers or a particular storage implementation.
// Service adds backend-neutral persistence and retrieval through BundleStore.
package diagnostics
