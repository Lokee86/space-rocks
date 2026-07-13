"""Validation for the observability schema, field, and event contracts."""

from __future__ import annotations

import re
from collections.abc import Iterable
from pathlib import Path

from data_sync.model.observability import ObservabilityContract
from data_sync import observability_toml


FIELD_TYPES = frozenset({"string", "integer", "number", "boolean", "uuid", "object"})
SENSITIVITIES = frozenset({"internal", "personal", "confidential", "restricted"})
RETENTION_TIERS = frozenset({"ephemeral_dev", "operational", "diagnostic_report", "audit_grade"})
SNAKE_CASE = re.compile(r"^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$")


class ObservabilityValidationError(Exception):
    """Raised when observability contract validation accumulates errors."""

    def __init__(self, errors: list[str]) -> None:
        self.errors = errors
        super().__init__("\n".join(errors))


def validate_observability(paths: Iterable[Path | str]) -> ObservabilityContract:
    """Load and validate observability schema, fields, and events."""
    resolved_paths = tuple(Path(path) for path in paths)
    contract = observability_toml.load_observability_contract(resolved_paths)
    errors: list[str] = []
    _validate_schema(contract, resolved_paths, errors)
    _validate_fields(contract, errors)
    _validate_events(contract, errors)
    _validate_redaction(contract, errors)
    _validate_retention(contract, errors)
    _validate_diagnostic_bundle(contract, errors)
    if errors:
        raise ObservabilityValidationError(errors)
    return contract


def _validate_schema(
    contract: ObservabilityContract, paths: tuple[Path, ...], errors: list[str]
) -> None:
    schema = contract.schema
    documents = {
        path.name: observability_toml._load_document(path)
        for path in paths
    }
    versions = (
        ("schema.toml.schema_version", documents["schema.toml"].get("schema_version")),
        ("events.toml.schema_version", documents["events.toml"].get("schema_version")),
        ("redaction.toml.schema_version", documents["redaction.toml"].get("schema_version")),
        (
            "retention_tiers.toml.schema_version",
            documents["retention_tiers.toml"].get("schema_version"),
        ),
        (
            "diagnostic_bundle.toml.schema_version",
            documents["diagnostic_bundle.toml"].get("schema_version"),
        ),
        (
            "diagnostic_bundle.toml.bundle.schema_version",
            documents["diagnostic_bundle.toml"].get("bundle", {}).get("schema_version"),
        ),
    )
    if len({version for _, version in versions}) != 1:
        errors.append(f"schema versions must agree: {versions!r}")

    limits = (
        ("schema.limits.max_batch_events", schema.limits.max_batch_events),
        ("schema.limits.max_event_bytes", schema.limits.max_event_bytes),
        ("schema.limits.max_event_fields", schema.limits.max_event_fields),
        ("schema.limits.max_string_bytes", schema.limits.max_string_bytes),
        ("schema.limits.max_collection_items", schema.limits.max_collection_items),
        ("schema.limits.max_free_form_fields", schema.limits.max_free_form_fields),
        ("schema.limits.max_free_form_value_bytes", schema.limits.max_free_form_value_bytes),
        ("diagnostic_bundle.max_events", contract.diagnostic_bundle.max_events),
        ("diagnostic_bundle.max_event_bytes", contract.diagnostic_bundle.max_event_bytes),
        ("diagnostic_bundle.max_total_bytes", contract.diagnostic_bundle.max_total_bytes),
        ("diagnostic_bundle.max_metadata_fields", contract.diagnostic_bundle.max_metadata_fields),
        ("diagnostic_bundle.max_correlation_fields", contract.diagnostic_bundle.max_correlation_fields),
        ("diagnostic_bundle.max_redaction_summary_entries", contract.diagnostic_bundle.max_redaction_summary_entries),
    )
    for name, value in limits:
        if value <= 0:
            errors.append(f"{name} must be positive: {value}")

    for name, pattern in (
        ("schema.envelope.event_name_pattern", schema.envelope.event_name_pattern),
        ("schema.identifiers.uuid_pattern", schema.identifiers.uuid_pattern),
        ("schema.free_form_fields.key_pattern", schema.free_form_fields.key_pattern),
    ):
        try:
            re.compile(pattern)
        except re.error as exc:
            errors.append(f"{name} is not a valid regex: {exc}")

    _unique_nonempty(schema.envelope.canonical_levels, "canonical level", errors)
    _unique_nonempty(schema.envelope.required_fields, "required field", errors)
    _unique_nonempty(schema.envelope.allowed_top_level_fields, "allowed top-level field", errors)


def _validate_fields(contract: ObservabilityContract, errors: list[str]) -> None:
    names = tuple(field.name for field in contract.fields)
    _unique_nonempty(names, "field definition", errors)
    for index, field in enumerate(contract.fields):
        label = f"field {field.name or index}"
        if not SNAKE_CASE.fullmatch(field.name):
            errors.append(f"{label} name must be snake_case: {field.name!r}")
        if field.type not in FIELD_TYPES:
            errors.append(f"{label} has unsupported type: {field.type}")
        if field.sensitivity not in SENSITIVITIES:
            errors.append(f"{label} has unsupported sensitivity: {field.sensitivity}")

    required = tuple(field.name for field in contract.fields if field.required)
    if set(contract.schema.envelope.required_fields) != set(required):
        errors.append(
            "required_fields does not exactly match required field definitions: "
            f"declared={sorted(contract.schema.envelope.required_fields)!r}, "
            f"defined={sorted(required)!r}"
        )

    catalog = set(names)
    allowed = set(contract.schema.envelope.allowed_top_level_fields)
    if allowed != catalog:
        errors.append(
            "allowed_top_level_fields does not exactly match field catalog: "
            f"declared={sorted(allowed)!r}, catalog={sorted(catalog)!r}"
        )


def _validate_events(contract: ObservabilityContract, errors: list[str]) -> None:
    try:
        event_pattern = re.compile(contract.schema.envelope.event_name_pattern)
    except re.error:
        event_pattern = None

    _unique_nonempty((event.name for event in contract.events), "event", errors)
    canonical_levels = set(contract.schema.envelope.canonical_levels)
    for index, event in enumerate(contract.events):
        label = f"event {event.name or index}"
        if event_pattern is not None and event_pattern.fullmatch(event.name) is None:
            errors.append(f"{label} name does not match event regex: {event.name!r}")
        if not event.category or not SNAKE_CASE.fullmatch(event.category):
            errors.append(f"{label} category must be non-empty snake_case: {event.category!r}")
        if not event.services or any(not service or not SNAKE_CASE.fullmatch(service) for service in event.services):
            errors.append(f"{label} services must be non-empty snake_case names: {event.services!r}")
        if event.default_level not in canonical_levels:
            errors.append(f"{label} default level is not canonical: {event.default_level!r}")
        if event.retention_tier not in {tier.name for tier in contract.retention_tiers}:
            errors.append(f"{label} retention tier is not declared: {event.retention_tier!r}")


def _validate_redaction(contract: ObservabilityContract, errors: list[str]) -> None:
    policy = contract.redaction
    action_names = tuple(action.name for action in policy.actions)
    _unique_nonempty(action_names, "redaction action name", errors)
    declared_actions = set(action_names)

    for action in policy.actions:
        if action.name.casefold() in {"reject", "redact"} and action.value_preserved:
            errors.append(
                f"redaction action {action.name!r} must not preserve unsafe values"
            )

    if policy.rejected_values_must_be_preserved:
        errors.append("redaction policy must not preserve rejected values")
    if policy.raw_forbidden_values_may_be_logged:
        errors.append("redaction policy must not log rejected raw values")
    if not policy.rejected_content_must_be_discarded:
        errors.append("redaction policy must discard rejected content")

    exact_rules: set[tuple[str, str, tuple[str, ...]]] = set()
    for index, rule in enumerate(policy.exact_rules):
        label = f"redaction exact rule {rule.category or index}"
        if not rule.category:
            errors.append(f"{label} category must be non-empty")
        _nonempty_unique_casefold(rule.keys, f"{label} key", errors)
        _declared_action(rule.action, label, declared_actions, errors)
        identity = (rule.category, rule.action, tuple(sorted(key.casefold() for key in rule.keys)))
        if identity in exact_rules:
            errors.append(f"duplicate redaction exact rule: {identity!r}")
        exact_rules.add(identity)

    fragment_rules: set[tuple[str, str, tuple[str, ...]]] = set()
    for index, rule in enumerate(policy.fragment_rules):
        label = f"redaction fragment rule {rule.category or index}"
        if not rule.category:
            errors.append(f"{label} category must be non-empty")
        _nonempty_unique_casefold(rule.fragments, f"{label} fragment", errors)
        _declared_action(rule.action, label, declared_actions, errors)
        identity = (rule.category, rule.action, tuple(sorted(fragment.casefold() for fragment in rule.fragments)))
        if identity in fragment_rules:
            errors.append(f"duplicate redaction fragment rule: {identity!r}")
        fragment_rules.add(identity)

    for label, action in (
        ("missing_action", policy.missing_action),
        ("ambiguous_match_action", policy.ambiguous_match_action),
        ("redaction_failure_action", policy.redaction_failure_action),
    ):
        _declared_action(action, f"redaction validation {label}", declared_actions, errors)


def _validate_retention(contract: ObservabilityContract, errors: list[str]) -> None:
    policy = contract.retention_policy
    if policy.production_legal_retention_commitment:
        errors.append("retention policy production/legal commitment must be false")
    if policy.unset_age_value < 0:
        errors.append(f"retention policy unset age must be nonnegative: {policy.unset_age_value}")

    tier_names = tuple(tier.name for tier in contract.retention_tiers)
    _unique(tier_names, "retention tier", errors)
    if set(tier_names) != RETENTION_TIERS:
        errors.append(
            "retention tier names must be exactly "
            f"{sorted(RETENTION_TIERS)!r}: {sorted(set(tier_names))!r}"
        )
    for index, tier in enumerate(contract.retention_tiers):
        label = f"retention tier {tier.name or index}"
        if tier.default_age_seconds < 0:
            errors.append(f"{label} default age must be nonnegative: {tier.default_age_seconds}")
        if tier.max_age_seconds < 0:
            errors.append(f"{label} max age must be nonnegative: {tier.max_age_seconds}")
        if tier.max_age_seconds != 0 and tier.default_age_seconds > tier.max_age_seconds:
            errors.append(
                f"{label} default age must not exceed max age: "
                f"{tier.default_age_seconds} > {tier.max_age_seconds}"
            )
        if tier.legal_commitment:
            errors.append(f"{label} legal commitment must be false")


def _validate_diagnostic_bundle(contract: ObservabilityContract, errors: list[str]) -> None:
    bundle = contract.diagnostic_bundle
    if bundle.diagnostic_report_id_type != "UUID":
        errors.append(
            "diagnostic_bundle diagnostic_report_id_type must be UUID: "
            f"{bundle.diagnostic_report_id_type!r}"
        )
    if not bundle.diagnostic_report_id_required:
        errors.append("diagnostic_bundle diagnostic_report_id must be required at bundle level")
    for section in (bundle.metadata, bundle.correlation, bundle.redaction_summary):
        if not section.required:
            errors.append(f"diagnostic_bundle {section.name} section must be required")

    canonical = {field.name: field for field in contract.fields}
    allowed_names = tuple(field.name for field in bundle.allowed_event_fields)
    _unique_nonempty(allowed_names, "diagnostic event field", errors)
    for field in bundle.allowed_event_fields:
        label = f"diagnostic event field {field.name or '<empty>'}"
        canonical_field = canonical.get(field.name)
        if canonical_field is None:
            errors.append(f"{label} is not in canonical field catalog")
        elif field.type != canonical_field.type:
            errors.append(
                f"{label} type does not match canonical field: "
                f"{field.type!r} != {canonical_field.type!r}"
            )

    diagnostic_required = {
        field.name for field in bundle.allowed_event_fields if field.required
    }
    canonical_required = set(contract.schema.envelope.required_fields)
    excess = sorted(diagnostic_required - canonical_required - {"diagnostic_report_id"})
    if excess:
        errors.append(
            "diagnostic required event fields are stricter than canonical envelope: "
            f"{excess!r}"
        )
    if "diagnostic_report_id" in diagnostic_required:
        errors.append(
            "diagnostic_report_id must remain bundle-level and must not be a required event field"
        )

    for name, value in (
        ("diagnostic_bundle.max_events", bundle.max_events),
        ("diagnostic_bundle.max_event_bytes", bundle.max_event_bytes),
        ("diagnostic_bundle.max_total_bytes", bundle.max_total_bytes),
        ("diagnostic_bundle.max_request_bytes", bundle.max_request_bytes),
        ("diagnostic_bundle.max_embedded_event_count", bundle.max_embedded_event_count),
        ("diagnostic_bundle.max_user_description_bytes", bundle.max_user_description_bytes),
        ("diagnostic_bundle.max_embedded_event_message_bytes", bundle.max_embedded_event_message_bytes),
        ("diagnostic_bundle.max_metadata_fields", bundle.max_metadata_fields),
        ("diagnostic_bundle.max_correlation_fields", bundle.max_correlation_fields),
        ("diagnostic_bundle.max_redaction_summary_entries", bundle.max_redaction_summary_entries),
    ):
        if value <= 0:
            errors.append(f"{name} must be positive: {value}")
    if bundle.max_request_bytes != bundle.max_total_bytes:
        errors.append("diagnostic_bundle.max_request_bytes must match max_total_bytes")
    if bundle.max_embedded_event_count != bundle.max_events:
        errors.append("diagnostic_bundle.max_embedded_event_count must match max_events")
    if bundle.max_user_description_bytes != contract.schema.limits.max_string_bytes:
        errors.append("diagnostic_bundle.max_user_description_bytes must match schema.limits.max_string_bytes")
    if bundle.max_embedded_event_message_bytes != contract.schema.limits.max_string_bytes:
        errors.append("diagnostic_bundle.max_embedded_event_message_bytes must match schema.limits.max_string_bytes")
    _nonempty_unique_casefold(bundle.allowed_triggers, "diagnostic trigger", errors)
    for trigger in bundle.allowed_triggers:
        if not SNAKE_CASE.fullmatch(trigger):
            errors.append(f"diagnostic trigger must be snake_case: {trigger!r}")

    markers = {
        action.replacement_marker
        for action in contract.redaction.actions
        if action.replacement_used and action.replacement_marker is not None
    }
    for action in contract.redaction.actions:
        if action.replacement_used and not action.replacement_marker:
            errors.append(f"redaction action {action.name!r} requires a replacement marker")
        if not action.replacement_used and action.replacement_marker is not None:
            errors.append(
                f"redaction action {action.name!r} must not define a replacement marker"
            )
    if len(markers) != 1:
        errors.append(f"redaction replacement markers must agree: {sorted(markers)!r}")
    elif bundle.replacement_marker != next(iter(markers)):
        errors.append(
            "diagnostic replacement marker does not match redaction marker: "
            f"{bundle.replacement_marker!r} != {next(iter(markers))!r}"
        )
    if bundle.raw_value_preserved or bundle.rejected_event_included or bundle.rejected_field_included:
        errors.append("diagnostic bundle must not preserve or include rejected/raw forbidden values")


def _declared_action(
    action: str, label: str, declared_actions: set[str], errors: list[str]
) -> None:
    if action not in declared_actions:
        errors.append(f"{label} references undeclared redaction action: {action!r}")


def _nonempty_unique_casefold(
    values: Iterable[str], label: str, errors: list[str]
) -> None:
    values = tuple(values)
    if not values:
        errors.append(f"{label} list must be non-empty")
    seen: set[str] = set()
    for value in values:
        if not value:
            errors.append(f"{label} must be non-empty")
        folded = value.casefold()
        if folded in seen:
            errors.append(f"duplicate {label} (case-insensitive): {value}")
        seen.add(folded)


def _unique(values: Iterable[str], label: str, errors: list[str]) -> None:
    seen: set[str] = set()
    for value in values:
        if value in seen:
            errors.append(f"duplicate {label}: {value}")
        seen.add(value)


def _unique_nonempty(values: Iterable[str], label: str, errors: list[str]) -> None:
    values = tuple(values)
    for value in values:
        if not value:
            errors.append(f"{label} must be non-empty")
    _unique(values, label, errors)
