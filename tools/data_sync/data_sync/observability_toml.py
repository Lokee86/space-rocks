"""Observability TOML source loading."""

from __future__ import annotations

from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Any, Iterable

from data_sync.model.observability import (
    DiagnosticBundle,
    DiagnosticBundleEventField,
    DiagnosticBundleSection,
    ObservabilityContract,
    ObservabilityEvent,
    ObservabilityFileLoggingDefaults,
    ObservabilityField,
    ObservabilityFreeForm,
    ObservabilityIdentifiers,
    ObservabilityLimits,
    ObservabilityRedactionAction,
    ObservabilityRedactionExactRule,
    ObservabilityRedactionFragmentRule,
    ObservabilityRedactionPolicy,
    ObservabilityRetentionPolicy,
    ObservabilityRetentionTier,
    ObservabilitySchema,
    ObservabilityValidation,
    ObservabilityEnvelope,
)


SOURCE_KINDS = {
    "schema.toml": "schema",
    "events.toml": "events",
    "fields.toml": "fields",
    "redaction.toml": "redaction",
    "retention_tiers.toml": "retention_tiers",
    "diagnostic_bundle.toml": "diagnostic_bundle",
}


class ObservabilityTomlError(Exception):
    """Raised when observability TOML cannot be loaded structurally."""


def load_observability_contract(paths: Iterable[Path | str]) -> ObservabilityContract:
    """Load the six configured observability TOML sources in source order."""
    resolved_paths = tuple(Path(path) for path in paths)
    if not resolved_paths:
        raise ObservabilityTomlError("observability TOML paths must not be empty")

    by_kind: dict[str, Path] = {}
    for path in resolved_paths:
        kind = SOURCE_KINDS.get(path.name)
        if kind is None:
            raise ObservabilityTomlError(
                f"{path}: filename must be one of: {', '.join(SOURCE_KINDS)}"
            )
        if kind in by_kind:
            raise ObservabilityTomlError(
                f"{path}: duplicate observability source kind {kind!r}; "
                f"already provided by {by_kind[kind]}"
            )
        by_kind[kind] = path

    missing = [kind for kind in SOURCE_KINDS.values() if kind not in by_kind]
    if missing:
        raise ObservabilityTomlError(
            "missing observability source kind(s): " + ", ".join(missing)
        )

    documents = {kind: _load_document(path) for kind, path in by_kind.items()}
    return ObservabilityContract(
        schema=_schema(documents["schema"], by_kind["schema"]),
        fields=_fields(documents["fields"], by_kind["fields"]),
        events=_events(documents["events"], by_kind["events"]),
        redaction=_redaction(documents["redaction"], by_kind["redaction"]),
        retention_policy=_retention_policy(
            documents["retention_tiers"], by_kind["retention_tiers"]
        ),
        file_logging=_file_logging(
            documents["retention_tiers"], by_kind["retention_tiers"]
        ),
        retention_tiers=_retention_tiers(
            documents["retention_tiers"], by_kind["retention_tiers"]
        ),
        diagnostic_bundle=_diagnostic_bundle(
            documents["diagnostic_bundle"], by_kind["diagnostic_bundle"]
        ),
    )


def load_observability_toml(paths: Iterable[Path | str]) -> ObservabilityContract:
    """Compatibility name for loading the observability contract sources."""
    return load_observability_contract(paths)


def _schema(document: Mapping[str, Any], path: Path) -> ObservabilitySchema:
    envelope = _table(document, "envelope", path)
    identifiers = _table(document, "identifiers", path)
    limits = _table(document, "limits", path)
    free_form = _table(document, "free_form_fields", path)
    validation = _table(document, "validation", path)
    return ObservabilitySchema(
        schema_version=_int(document, "schema_version", path),
        envelope=ObservabilityEnvelope(
            canonical_levels=_strings(envelope, "canonical_levels", path, "envelope"),
            required_fields=_strings(envelope, "required_fields", path, "envelope"),
            timestamp_format=_string(envelope, "timestamp_format", path, "envelope"),
            event_name_pattern=_string(envelope, "event_name_pattern", path, "envelope"),
            reject_unknown_top_level_fields=_bool(envelope, "reject_unknown_top_level_fields", path, "envelope"),
            allowed_top_level_fields=_strings(envelope, "allowed_top_level_fields", path, "envelope"),
        ),
        identifiers=ObservabilityIdentifiers(
            new_id_format=_string(identifiers, "new_id_format", path, "identifiers"),
            new_id_version=_string(identifiers, "new_id_version", path, "identifiers"),
            uuid_pattern=_string(identifiers, "uuid_pattern", path, "identifiers"),
        ),
        limits=ObservabilityLimits(
            max_batch_events=_int(limits, "max_batch_events", path, "limits"),
            max_event_bytes=_int(limits, "max_event_bytes", path, "limits"),
            max_event_fields=_int(limits, "max_event_fields", path, "limits"),
            max_string_bytes=_int(limits, "max_string_bytes", path, "limits"),
            max_collection_items=_int(limits, "max_collection_items", path, "limits"),
            max_free_form_fields=_int(limits, "max_free_form_fields", path, "limits"),
            max_free_form_value_bytes=_int(limits, "max_free_form_value_bytes", path, "limits"),
        ),
        free_form_fields=ObservabilityFreeForm(
            key_pattern=_string(free_form, "key_pattern", path, "free_form_fields"),
            value_types=_strings(free_form, "value_types", path, "free_form_fields"),
            values_must_be_redacted=_bool(free_form, "values_must_be_redacted", path, "free_form_fields"),
        ),
        validation=ObservabilityValidation(
            unknown_top_level_field_action=_string(validation, "unknown_top_level_field_action", path, "validation"),
            unsafe_field_action=_string(validation, "unsafe_field_action", path, "validation"),
            null_values_allowed=_bool(validation, "null_values_allowed", path, "validation"),
        ),
    )


def _fields(document: Mapping[str, Any], path: Path) -> tuple[ObservabilityField, ...]:
    return tuple(
        ObservabilityField(
            name=_string(item, "name", path, f"fields[{index}]"),
            type=_string(item, "type", path, f"fields[{index}]"),
            required=_bool(item, "required", path, f"fields[{index}]"),
            description=_string(item, "description", path, f"fields[{index}]"),
            sensitivity=_string(item, "sensitivity", path, f"fields[{index}]"),
        )
        for index, item in enumerate(_list(document, "fields", path))
    )


def _events(document: Mapping[str, Any], path: Path) -> tuple[ObservabilityEvent, ...]:
    return tuple(
        ObservabilityEvent(
            name=_string(item, "name", path, f"events[{index}]"),
            category=_string(item, "category", path, f"events[{index}]"),
            default_level=_string(item, "default_level", path, f"events[{index}]"),
            description=_string(item, "description", path, f"events[{index}]"),
            services=_strings(item, "services", path, f"events[{index}]"),
            trace_required=_bool(item, "trace_required", path, f"events[{index}]"),
            audit_eligible=_bool(item, "audit_eligible", path, f"events[{index}]"),
            retention_tier=_string(item, "retention_tier", path, f"events[{index}]"),
        )
        for index, item in enumerate(_list(document, "events", path))
    )


def _redaction(document: Mapping[str, Any], path: Path) -> ObservabilityRedactionPolicy:
    policy = _table(document, "policy", path)
    validation = _table(document, "validation", path)
    actions = tuple(
        ObservabilityRedactionAction(
            name=_string(item, "name", path, f"action.{name}"),
            description=_string(item, "description", path, f"action.{name}"),
            value_preserved=_bool(item, "value_preserved", path, f"action.{name}"),
            replacement_used=_bool(item, "replacement_used", path, f"action.{name}"),
            replacement_marker=_optional_string(item, "replacement_marker", path, f"action.{name}"),
        )
        for name, item in _named_tables(document, "action", path)
    )
    exact_rules = tuple(
        ObservabilityRedactionExactRule(
            category=_string(item, "category", path, f"forbidden_exact_keys[{index}]"),
            action=_string(item, "action", path, f"forbidden_exact_keys[{index}]"),
            keys=_strings(item, "keys", path, f"forbidden_exact_keys[{index}]"),
        )
        for index, item in enumerate(_list(document, "forbidden_exact_keys", path))
    )
    fragment_rules = tuple(
        ObservabilityRedactionFragmentRule(
            category=_string(item, "category", path, f"forbidden_key_fragments[{index}]"),
            action=_string(item, "action", path, f"forbidden_key_fragments[{index}]"),
            fragments=_strings(item, "fragments", path, f"forbidden_key_fragments[{index}]"),
        )
        for index, item in enumerate(_list(document, "forbidden_key_fragments", path))
    )
    return ObservabilityRedactionPolicy(
        schema_version=_int(document, "schema_version", path),
        case_sensitive=_bool(policy, "case_sensitive", path, "policy"),
        key_matching=_string(policy, "key_matching", path, "policy"),
        forbidden_value_handling=_string(policy, "forbidden_value_handling", path, "policy"),
        rejected_values_must_be_preserved=_bool(policy, "rejected_values_must_be_preserved", path, "policy"),
        raw_forbidden_values_may_be_logged=_bool(policy, "raw_forbidden_values_may_be_logged", path, "policy"),
        actions=actions,
        exact_rules=exact_rules,
        fragment_rules=fragment_rules,
        missing_action=_string(validation, "missing_action", path, "validation"),
        ambiguous_match_action=_string(validation, "ambiguous_match_action", path, "validation"),
        redaction_failure_action=_string(validation, "redaction_failure_action", path, "validation"),
        rejected_content_must_be_discarded=_bool(validation, "rejected_content_must_be_discarded", path, "validation"),
    )


def _retention_policy(document: Mapping[str, Any], path: Path) -> ObservabilityRetentionPolicy:
    policy = _table(document, "policy", path)
    return ObservabilityRetentionPolicy(
        age_values_are_configuration_metadata=_bool(policy, "age_values_are_configuration_metadata", path, "policy"),
        production_legal_retention_commitment=_bool(policy, "production_legal_retention_commitment", path, "policy"),
        unset_age_value=_int(policy, "unset_age_value", path, "policy"),
        age_unit=_string(policy, "age_unit", path, "policy"),
    )


def _retention_tiers(document: Mapping[str, Any], path: Path) -> tuple[ObservabilityRetentionTier, ...]:
    return tuple(
        ObservabilityRetentionTier(
            name=_string(item, "name", path, f"tiers[{index}]"),
            purpose=_string(item, "purpose", path, f"tiers[{index}]"),
            durability=_string(item, "durability", path, f"tiers[{index}]"),
            compression=_string(item, "compression", path, f"tiers[{index}]"),
            delete_policy=_string(item, "delete_policy", path, f"tiers[{index}]"),
            default_age_seconds=_int(item, "default_age_seconds", path, f"tiers[{index}]"),
            max_age_seconds=_int(item, "max_age_seconds", path, f"tiers[{index}]"),
            age_is_configurable=_bool(item, "age_is_configurable", path, f"tiers[{index}]"),
            legal_commitment=_bool(item, "legal_commitment", path, f"tiers[{index}]"),
        )
        for index, item in enumerate(_list(document, "tiers", path))
    )


def _file_logging(document: Mapping[str, Any], path: Path) -> ObservabilityFileLoggingDefaults:
    defaults = _table(document, "file_logging", path)
    return ObservabilityFileLoggingDefaults(
        max_active_segment_age_seconds=_int(
            defaults, "max_active_segment_age_seconds", path, "file_logging"
        ),
        compression_enabled=_bool(defaults, "compression_enabled", path, "file_logging"),
    )


def _diagnostic_bundle(document: Mapping[str, Any], path: Path) -> DiagnosticBundle:
    bundle = _table(document, "bundle", path)
    limits = _table(document, "limits", path)
    metadata, correlation, summary = (
        _table(_table(document, "sections", path), name, path, "sections")
        for name in ("metadata", "correlation", "redaction_summary")
    )
    events = _table(document, "events", path)
    forbidden = _table(document, "forbidden_content", path)
    section = lambda table, name: DiagnosticBundleSection(
        name=name,
        required=_bool(table, "required", path, f"sections.{name}"),
        allowed_fields=_strings(table, "allowed_fields", path, f"sections.{name}"),
    )
    allowed = tuple(
        DiagnosticBundleEventField(
            name=_string(item, "name", path, f"allowed_event_fields[{index}]"),
            type=_string(item, "type", path, f"allowed_event_fields[{index}]"),
            required=_bool(item, "required", path, f"allowed_event_fields[{index}]"),
        )
        for index, item in enumerate(_list(document, "allowed_event_fields", path))
    )
    return DiagnosticBundle(
        name=_string(bundle, "name", path, "bundle"),
        schema_version=_int(bundle, "schema_version", path, "bundle"),
        diagnostic_report_id_type=_string(bundle, "diagnostic_report_id_type", path, "bundle"),
        diagnostic_report_id_required=_bool(bundle, "diagnostic_report_id_required", path, "bundle"),
        safe_to_copy_default=_bool(bundle, "safe_to_copy_default", path, "bundle"),
        safe_to_copy_field=_string(bundle, "safe_to_copy_field", path, "bundle"),
        max_events=_int(limits, "max_events", path, "limits"),
        max_event_bytes=_int(limits, "max_event_bytes", path, "limits"),
        max_total_bytes=_int(limits, "max_total_bytes", path, "limits"),
        max_request_bytes=_int(limits, "max_request_bytes", path, "limits"),
        max_embedded_event_count=_int(limits, "max_embedded_event_count", path, "limits"),
        max_user_description_bytes=_int(limits, "max_user_description_bytes", path, "limits"),
        max_embedded_event_message_bytes=_int(limits, "max_embedded_event_message_bytes", path, "limits"),
        max_metadata_fields=_int(limits, "max_metadata_fields", path, "limits"),
        max_correlation_fields=_int(limits, "max_correlation_fields", path, "limits"),
        max_redaction_summary_entries=_int(limits, "max_redaction_summary_entries", path, "limits"),
        allowed_triggers=_strings(_table(document, "trigger_policy", path), "allowed_triggers", path, "trigger_policy"),
        metadata=section(metadata, "metadata"),
        correlation=section(correlation, "correlation"),
        redaction_summary=section(summary, "redaction_summary"),
        events_required=_bool(events, "required", path, "events"),
        events_ordered_by=_string(events, "ordered_by", path, "events"),
        truncated_when_limit_exceeded=_bool(events, "truncated_when_limit_exceeded", path, "events"),
        allowed_event_fields=allowed,
        forbidden_content_behavior=_string(forbidden, "behavior", path, "forbidden_content"),
        raw_value_preserved=_bool(forbidden, "raw_value_preserved", path, "forbidden_content"),
        rejected_event_included=_bool(forbidden, "rejected_event_included", path, "forbidden_content"),
        rejected_field_included=_bool(forbidden, "rejected_field_included", path, "forbidden_content"),
        replacement_marker=_string(forbidden, "replacement_marker", path, "forbidden_content"),
        redaction_summary_must_count_rejected_content=_bool(forbidden, "redaction_summary_must_count_rejected_content", path, "forbidden_content"),
        redaction_failure_behavior=_string(forbidden, "redaction_failure_behavior", path, "forbidden_content"),
        unknown_event_field_behavior=_string(forbidden, "unknown_event_field_behavior", path, "forbidden_content"),
    )


def _load_document(path: Path) -> Mapping[str, Any]:
    try:
        text = path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise ObservabilityTomlError(f"{path}: file does not exist") from exc
    except OSError as exc:
        raise ObservabilityTomlError(f"{path}: failed to read TOML: {exc}") from exc
    try:
        return _toml_loads(text)
    except Exception as exc:
        raise ObservabilityTomlError(f"{path}: failed to parse TOML: {exc}") from exc


def _toml_loads(text: str) -> Mapping[str, Any]:
    try:
        import tomllib
    except ModuleNotFoundError:
        import tomli as tomllib
    return tomllib.loads(text)


def _table(document: Mapping[str, Any], key: str, path: Path, parent: str | None = None) -> Mapping[str, Any]:
    label = f"{parent}.{key}" if parent else key
    value = _value(document, key, path, label)
    if not isinstance(value, Mapping):
        raise ObservabilityTomlError(f"{path}: {label} must be a table")
    return value


def _list(document: Mapping[str, Any], key: str, path: Path) -> Sequence[Any]:
    value = _value(document, key, path, key)
    if not isinstance(value, Sequence) or isinstance(value, (str, bytes, bytearray)):
        raise ObservabilityTomlError(f"{path}: {key} must be a list")
    return value


def _named_tables(document: Mapping[str, Any], key: str, path: Path) -> Iterable[tuple[str, Mapping[str, Any]]]:
    value = _value(document, key, path, key)
    if not isinstance(value, Mapping):
        raise ObservabilityTomlError(f"{path}: {key} must be a table")
    for name, item in value.items():
        if not isinstance(item, Mapping):
            raise ObservabilityTomlError(f"{path}: {key}.{name} must be a table")
        yield name, item


def _value(document: Mapping[str, Any], key: str, path: Path, label: str) -> Any:
    if key not in document:
        raise ObservabilityTomlError(f"{path}: missing key {label}")
    return document[key]


def _string(document: Mapping[str, Any], key: str, path: Path, parent: str | None = None) -> str:
    label = f"{parent}.{key}" if parent else key
    value = _value(document, key, path, label)
    if not isinstance(value, str):
        raise ObservabilityTomlError(f"{path}: {label} must be a string")
    return value


def _optional_string(document: Mapping[str, Any], key: str, path: Path, parent: str) -> str | None:
    if key not in document:
        return None
    return _string(document, key, path, parent)


def _int(document: Mapping[str, Any], key: str, path: Path, parent: str | None = None) -> int:
    label = f"{parent}.{key}" if parent else key
    value = _value(document, key, path, label)
    if not isinstance(value, int) or isinstance(value, bool):
        raise ObservabilityTomlError(f"{path}: {label} must be an integer")
    return value


def _bool(document: Mapping[str, Any], key: str, path: Path, parent: str | None = None) -> bool:
    label = f"{parent}.{key}" if parent else key
    value = _value(document, key, path, label)
    if not isinstance(value, bool):
        raise ObservabilityTomlError(f"{path}: {label} must be a boolean")
    return value


def _strings(document: Mapping[str, Any], key: str, path: Path, parent: str | None = None) -> tuple[str, ...]:
    label = f"{parent}.{key}" if parent else key
    value = _value(document, key, path, label)
    if not isinstance(value, Sequence) or isinstance(value, (str, bytes, bytearray)):
        raise ObservabilityTomlError(f"{path}: {label} must be a list")
    result = tuple(value)
    if not all(isinstance(item, str) for item in result):
        raise ObservabilityTomlError(f"{path}: {label} must be a list of strings")
    return result
