---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-711c-ab9e-81b3771987f6
document_type: general
policy_exempt: false
summary: Observability is an active tools/datasync source-of-truth domain. The editable contract is the TOML source set below. data-sync validates the cross-file model and generates all language/runtime metadata and the generated contract...
---
# Observability Contract

Parent index: [Data](./!INDEX.md)

## Purpose and ownership

Observability is an active `tools/data_sync` source-of-truth domain. The editable contract is the TOML source set below. `data-sync` validates the cross-file model and generates all language/runtime metadata and the generated contract reference. Runtime emitters consume generated metadata; generated outputs are never edited manually.

## Source files

```text
shared/contracts/observability/schema.toml
shared/contracts/observability/services.toml
shared/contracts/observability/events.toml
shared/contracts/observability/fields.toml
shared/contracts/observability/redaction.toml
shared/contracts/observability/retention_tiers.toml
shared/contracts/observability/diagnostic_bundle.toml
```

Together these files define the envelope, service registry, event eligibility and defaults, context/free-form fields, redaction policy, retention tiers, and diagnostic-bundle contract. No service may add a local event definition or policy that changes these generated rules.

## Generated outputs

```text
shared/go/observabilityevent/contract_generated.go
client/scripts/generated/observability/contract_generated.gd
services/api-server/app/lib/observability/contract_generated.rb
shared/contracts/observability/generated/contract.json
docs/observability/generated/contract-reference.md
```

The generated Markdown reference is a read-only rendered reference. It changes only through the observability data-sync generator.

## Commands

From the repository root:

```bash
data-sync -validate -observability
data-sync -diff -observability -go -gds -ruby -json -docs
data-sync -push -observability -go -gds -ruby -json -docs
data-sync -check -observability -go -gds -ruby -json -docs
```

`-validate` reads and cross-validates the source contract without writing. `-diff` renders every selected output and reports drift. `-push` writes selected generated outputs. `-check` renders without writing and fails when any selected output differs.

## Validation rules

Validation covers:

- one valid schema/envelope definition and supported schema version;
- unique service keys and emitted names;
- unique event names, categories, levels, service allowlists, trace requirements, and bridge-only declarations;
- context field names, types, UUID requirements, and limits;
- free-form key/value/count/size limits;
- redaction exact/fragment rules and rejection of ambiguous actions;
- retention tier references and file policy values;
- diagnostic-bundle field/event references;
- deterministic output ordering and complete configured target files.

Generation fails rather than silently dropping an unsupported language or output. The Ruby selector is observability-specific; JSON and docs selectors apply to observability as well as realtime wire.

## Generation ownership and consumers

`tools/data_sync/config.toml` owns the `[sot.observability]` source list and the `observability.go`, `observability.gds`, `observability.ruby`, `observability.json`, and `observability.docs` targets. The tool implementation and generators are the only owners of generated output content.

Consumers are:

- Go services through `shared/go/observabilityevent` and service `logging.Emit` boundaries;
- the Godot client through `observability_emitter.gd` and generated GDScript constants;
- API Rails through `Observability::Emitter`, `WorkerRuntime`, and `PumaHooks`;
- diagnostic-aggregator through generated Go observability metadata;
- handwritten service/data docs through this document and the generated contract reference.

## Failure modes

- malformed TOML or a missing source file fails validation;
- duplicate service/event/field/retention identifiers fail validation;
- an event referencing an unknown service, field, or retention tier fails validation;
- invalid trace/UUID, field type, redaction, or limit declarations fail validation;
- unsupported selector combinations fail CLI parsing;
- generated output drift causes `-check` to exit nonzero;
- a runtime emitter rejects an event, unsafe value, or oversized record according to generated rejection codes;
- a writer failure degrades the service-owned runtime but must not cause the workflow to raise or serialize the rejected/private payload.

The contract pipeline and runtime writer failures are separate: data-sync failure blocks a generated-contract change; a runtime write failure is a bounded service degradation.

## Rule against manual edits

Do not edit any generated output directly. Change the owning TOML source, run validation, review the diff, push the selected outputs, and run the check command. A manual edit to a generated Go, GDScript, Ruby, JSON, or Markdown file is drift and must be overwritten by the next push.

## Code and test map

```text
tools/data_sync/config.toml
tools/data_sync/main.py
tools/data_sync/data_sync/cli.py
tools/data_sync/data_sync/config.py
tools/data_sync/data_sync/observability_toml.py
tools/data_sync/data_sync/observability_validate.py
tools/data_sync/data_sync/observability_sync.py
tools/data_sync/data_sync/model/observability.py
tools/data_sync/data_sync/generators/observability_go.py
tools/data_sync/data_sync/generators/observability_gds.py
tools/data_sync/data_sync/generators/observability_ruby.py
tools/data_sync/data_sync/generators/observability_json.py
tools/data_sync/data_sync/generators/observability_docs.py
tools/data_sync/tests/test_observability_toml.py
tools/data_sync/tests/test_observability_validate.py
tools/data_sync/tests/test_observability_sync.py
tools/data_sync/tests/test_observability_generators.py
tools/data_sync/tests/test_observability_emission_guards.py
```

## Related docs

- [Data sync and SSoT pipeline](data-sync-and-ssot-pipeline.md)
- [Source-of-truth map](source-of-truth-map.md)
- [Canonical event emission](../observability/canonical-event-emission.md)
- [Generated contract reference](../observability/generated/contract-reference.md)
