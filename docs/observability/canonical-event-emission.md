# Canonical Event Emission

Parent index: [Observability](./!INDEX.md)

## Ownership

The canonical observability contract is owned by the source files under `shared/contracts/observability/`. The data-sync tool validates that contract and generates language metadata for Go, Ruby, and GDScript. Runtime emitters consume generated metadata; they do not maintain service names, event policy, limits, rejection codes, or redaction rules locally.

The five emitting components are:

- game server
- player data
- diagnostic aggregator
- API server
- client

All five write the same canonical JSONL envelope. Service adapters retain their existing public logging APIs and category/level controls, but canonical emitters own validation, redaction, serialization, event identity, and stable rejection outcomes.

## Runtime Boundaries

- `shared/go/observabilityevent` is the shared Go emitter used by the three Go services.
- `shared/go/servicelog` accepts already serialized canonical records through a narrow sink seam and must not reshape them through `slog`.
- `services/api-server/app/lib/observability/emitter.rb` is the API emitter. `StructuredFormatter` is its existing-logger compatibility adapter.
- `client/scripts/logging/observability_emitter.gd` is the client emitter. The client logger remains the category/level and console compatibility adapter.

Each emitter obtains the emitted service name from the generated service registry. Adapter-local emitted-name literals are forbidden.

## Legacy Logging Bridge

`log_message` is a bridge-only event for existing logging APIs and call sites. Ordinary canonical emission rejects it with `bridge_event_forbidden`. Only dedicated compatibility paths may emit it:

- Go `EmitLegacy` and `EmitLegacyArgs`
- Ruby `emit_legacy` through `StructuredFormatter`
- GDScript `emit_legacy` through the client logger

Legacy event labels are retained only as the scalar `fields.legacy_event` value. The bridge does not authorize new semantic call sites to use `log_message`, and it does not change ownership of domain behavior.

## Validation And Safety

Emitters enforce the generated contract before writing:

- event/service eligibility and trace requirements
- UUID and context-field validation
- free-form key, type, null, count, string, and event-size limits
- generated redaction and unsafe-field rejection policy
- canonical envelope serialization

Unsafe rejected values are not written to JSONL or warning output. Writer failures are non-raising and are exposed through emitter-owned status counters and stable `write_failed` results.

## Verification

### Diagnostic aggregator rollout

The diagnostic aggregator has completed its current-workflow canonical rollout. Hosted startup and shutdown own separate operation traces. HTTP requests own fresh request IDs and provisional traces, with valid submitted diagnostic traces continued after strict body decoding. The handler owns bounded rejection events; the report service owns accepted intake, successful durable storage through `diagnostic_report_stored`, and storage failure; the hosted service owns report-store shutdown failure.

No production file under `services/diagnostic-aggregator/` uses `EmitLegacy`, `EmitLegacyArgs`, or bridge-backed `Info`/`Error` calls. The repository guard is scoped to that service; `log_message` and its compatibility adapters remain available for services whose workflow rollout is incomplete.


Cross-language acceptance, redaction, and rejection behavior is exercised from:

```text
shared/contracts/observability/fixtures/emitter_cases.json
```

The client mirror is guarded byte-for-byte because Godot resources cannot read outside the project root. Repository guards also keep the bridge event confined to generated contract data, emitter internals, logging adapters, fixtures, and tests.

Generated outputs must remain drift-free:

```bash
python tools/data_sync/main.py -push -observability -go -gds -ruby -json -docs
python tools/data_sync/main.py -diff -observability -go -gds -ruby -json -docs
```

## Does Not Own

Canonical emission does not own diagnostic submission, metrics, tracing infrastructure, durable telemetry storage, gameplay behavior, networking behavior, or broad domain-event migration. Correlation identifiers are validated when supplied by their owning systems; emitters do not invent domain correlation policy.
