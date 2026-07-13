"""Concrete model for the observability source-of-truth contracts."""

from __future__ import annotations

from dataclasses import asdict, dataclass, is_dataclass
from typing import Any, Mapping


@dataclass(frozen=True)
class ObservabilityEnvelope:
    canonical_levels: tuple[str, ...]
    required_fields: tuple[str, ...]
    timestamp_format: str
    event_name_pattern: str
    reject_unknown_top_level_fields: bool
    allowed_top_level_fields: tuple[str, ...]


@dataclass(frozen=True)
class ObservabilityIdentifiers:
    new_id_format: str
    new_id_version: str
    uuid_pattern: str


@dataclass(frozen=True)
class ObservabilityLimits:
    max_batch_events: int
    max_event_bytes: int
    max_event_fields: int
    max_string_bytes: int
    max_collection_items: int
    max_free_form_fields: int
    max_free_form_value_bytes: int


@dataclass(frozen=True)
class ObservabilityFreeForm:
    key_pattern: str
    value_types: tuple[str, ...]
    values_must_be_redacted: bool


@dataclass(frozen=True)
class ObservabilityValidation:
    unknown_top_level_field_action: str
    unsafe_field_action: str
    null_values_allowed: bool


@dataclass(frozen=True)
class ObservabilitySchema:
    schema_version: int
    envelope: ObservabilityEnvelope
    identifiers: ObservabilityIdentifiers
    limits: ObservabilityLimits
    free_form_fields: ObservabilityFreeForm
    validation: ObservabilityValidation


@dataclass(frozen=True)
class ObservabilityField:
    name: str
    type: str
    required: bool
    description: str
    sensitivity: str


@dataclass(frozen=True)
class ObservabilityEvent:
    name: str
    category: str
    default_level: str
    description: str
    services: tuple[str, ...]
    trace_required: bool
    audit_eligible: bool
    retention_tier: str


@dataclass(frozen=True)
class ObservabilityRedactionAction:
    name: str
    description: str
    value_preserved: bool
    replacement_used: bool
    replacement_marker: str | None = None


@dataclass(frozen=True)
class ObservabilityRedactionExactRule:
    category: str
    action: str
    keys: tuple[str, ...]


@dataclass(frozen=True)
class ObservabilityRedactionFragmentRule:
    category: str
    action: str
    fragments: tuple[str, ...]


@dataclass(frozen=True)
class ObservabilityRedactionPolicy:
    schema_version: int
    case_sensitive: bool
    key_matching: str
    forbidden_value_handling: str
    rejected_values_must_be_preserved: bool
    raw_forbidden_values_may_be_logged: bool
    actions: tuple[ObservabilityRedactionAction, ...]
    exact_rules: tuple[ObservabilityRedactionExactRule, ...]
    fragment_rules: tuple[ObservabilityRedactionFragmentRule, ...]
    missing_action: str
    ambiguous_match_action: str
    redaction_failure_action: str
    rejected_content_must_be_discarded: bool


@dataclass(frozen=True)
class ObservabilityRetentionPolicy:
    age_values_are_configuration_metadata: bool
    production_legal_retention_commitment: bool
    unset_age_value: int
    age_unit: str


@dataclass(frozen=True)
class ObservabilityRetentionTier:
    name: str
    purpose: str
    durability: str
    compression: str
    delete_policy: str
    default_age_seconds: int
    max_age_seconds: int
    age_is_configurable: bool
    legal_commitment: bool


@dataclass(frozen=True)
class DiagnosticBundleSection:
    name: str
    required: bool
    allowed_fields: tuple[str, ...]


@dataclass(frozen=True)
class DiagnosticBundleEventField:
    name: str
    type: str
    required: bool


@dataclass(frozen=True)
class DiagnosticBundle:
    name: str
    schema_version: int
    diagnostic_report_id_type: str
    diagnostic_report_id_required: bool
    safe_to_copy_default: bool
    safe_to_copy_field: str
    max_events: int
    max_event_bytes: int
    max_total_bytes: int
    max_metadata_fields: int
    max_correlation_fields: int
    max_redaction_summary_entries: int
    metadata: DiagnosticBundleSection
    correlation: DiagnosticBundleSection
    redaction_summary: DiagnosticBundleSection
    events_required: bool
    events_ordered_by: str
    truncated_when_limit_exceeded: bool
    allowed_event_fields: tuple[DiagnosticBundleEventField, ...]
    forbidden_content_behavior: str
    raw_value_preserved: bool
    rejected_event_included: bool
    rejected_field_included: bool
    replacement_marker: str
    redaction_summary_must_count_rejected_content: bool
    redaction_failure_behavior: str
    unknown_event_field_behavior: str


@dataclass(frozen=True)
class ObservabilityContract:
    schema: ObservabilitySchema
    fields: tuple[ObservabilityField, ...]
    events: tuple[ObservabilityEvent, ...]
    redaction: ObservabilityRedactionPolicy
    retention_policy: ObservabilityRetentionPolicy
    retention_tiers: tuple[ObservabilityRetentionTier, ...]
    diagnostic_bundle: DiagnosticBundle

    def field(self, name: str) -> ObservabilityField:
        for field in self.fields:
            if field.name == name:
                return field
        raise KeyError(name)

    def event(self, name: str) -> ObservabilityEvent:
        for event in self.events:
            if event.name == name:
                return event
        raise KeyError(name)

    def retention_tier(self, name: str) -> ObservabilityRetentionTier:
        for tier in self.retention_tiers:
            if tier.name == name:
                return tier
        raise KeyError(name)

    def as_dict(self) -> Mapping[str, Any]:
        return _json_value(
            {
                "schema": self.schema,
                "fields": self.fields,
                "events": self.events,
                "redaction": self.redaction,
                "retention_tiers": {
                    "policy": self.retention_policy,
                    "tiers": self.retention_tiers,
                },
                "diagnostic_bundle": self.diagnostic_bundle,
            }
        )


def _json_value(value: Any) -> Any:
    if is_dataclass(value):
        return {key: _json_value(item) for key, item in asdict(value).items()}
    if isinstance(value, tuple):
        return [_json_value(item) for item in value]
    if isinstance(value, Mapping):
        return {key: _json_value(item) for key, item in value.items()}
    return value
