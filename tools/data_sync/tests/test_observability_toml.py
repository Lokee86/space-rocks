from __future__ import annotations

import shutil
from pathlib import Path

import pytest

from data_sync.observability_toml import ObservabilityTomlError, load_observability_contract


OBSERVABILITY_PATHS = (
    "shared/contracts/observability/schema.toml",
    "shared/contracts/observability/events.toml",
    "shared/contracts/observability/fields.toml",
    "shared/contracts/observability/redaction.toml",
    "shared/contracts/observability/retention_tiers.toml",
    "shared/contracts/observability/diagnostic_bundle.toml",
)


def copied_observability_paths(tmp_path: Path) -> tuple[Path, ...]:
    repo_root = Path(__file__).resolve().parents[3]
    paths: list[Path] = []
    for relative_path in OBSERVABILITY_PATHS:
        source = repo_root / relative_path
        destination = tmp_path / Path(relative_path).name
        shutil.copyfile(source, destination)
        paths.append(destination)
    return tuple(paths)


def test_loads_canonical_observability_contract(tmp_path: Path) -> None:
    contract = load_observability_contract(copied_observability_paths(tmp_path))

    assert contract.schema.schema_version == 1
    assert contract.schema.envelope.canonical_levels == (
        "debug",
        "info",
        "warn",
        "error",
        "critical",
    )
    assert len(contract.fields) == 47
    assert tuple(field.name for field in contract.fields[:3]) == ("timestamp", "level", "event")
    assert contract.fields[-1].name == "fields"
    assert len(contract.events) == 148
    assert tuple(event.name for event in contract.events[:3]) == (
        "service_starting",
        "service_started",
        "service_stopping",
    )
    assert contract.events[-1].name == "soak_degradation_detected"
    assert len(contract.redaction.actions) == 2
    assert tuple(action.name for action in contract.redaction.actions) == ("reject", "redact")
    assert len(contract.redaction.exact_rules) == 8
    assert len(contract.redaction.fragment_rules) == 8
    assert tuple(tier.name for tier in contract.retention_tiers) == (
        "ephemeral_dev",
        "operational",
        "diagnostic_report",
        "audit_grade",
    )
    assert len(contract.diagnostic_bundle.allowed_event_fields) == 25

    assert contract.field("player_id").sensitivity == "personal"
    assert contract.field("player_id").required is False
    assert contract.event("service_startup_failed").default_level == "critical"
    assert contract.event("service_startup_failed").trace_required is True
    assert contract.retention_tier("audit_grade").durability == "durable"
    assert contract.diagnostic_bundle.events_ordered_by == "timestamp_ascending"
    assert contract.diagnostic_bundle.max_request_bytes == 5242880
    assert contract.diagnostic_bundle.max_embedded_event_count == 500
    assert contract.diagnostic_bundle.max_user_description_bytes == 4096
    assert contract.diagnostic_bundle.max_embedded_event_message_bytes == 4096
    assert contract.diagnostic_bundle.allowed_triggers == (
        "manual_bug_report",
        "development_submission",
        "crash",
        "startup_failure",
        "unrecoverable_state",
        "recovery_exhausted",
    )
    assert contract.retention_tier("diagnostic_report").default_age_seconds == 1209600
    assert contract.retention_tier("operational").default_age_seconds == 1209600
    assert contract.file_logging.max_active_segment_age_seconds == 3600
    assert contract.file_logging.compression_enabled is True


def test_rejects_missing_source_kind(tmp_path: Path) -> None:
    paths = copied_observability_paths(tmp_path)

    with pytest.raises(ObservabilityTomlError, match=r"missing observability source kind\(s\): diagnostic_bundle"):
        load_observability_contract(paths[:-1])


def test_rejects_duplicate_source_kind(tmp_path: Path) -> None:
    paths = copied_observability_paths(tmp_path)
    duplicate_directory = tmp_path / "duplicate"
    duplicate_directory.mkdir()
    duplicate = duplicate_directory / "schema.toml"
    shutil.copyfile(paths[0], duplicate)

    with pytest.raises(ObservabilityTomlError, match="duplicate observability source kind 'schema'"):
        load_observability_contract((*paths, duplicate))


def test_rejects_wrong_structural_type(tmp_path: Path) -> None:
    paths = copied_observability_paths(tmp_path)
    schema = paths[0]
    schema.write_text(
        schema.read_text(encoding="utf-8").replace(
            'canonical_levels = ["debug", "info", "warn", "error", "critical"]',
            'canonical_levels = "info"',
        ),
        encoding="utf-8",
    )

    with pytest.raises(ObservabilityTomlError, match=r"envelope\.canonical_levels must be a list"):
        load_observability_contract(paths)


def test_rejects_missing_required_key(tmp_path: Path) -> None:
    paths = copied_observability_paths(tmp_path)
    schema = paths[0]
    schema.write_text(
        schema.read_text(encoding="utf-8").replace("schema_version = 1\n", "", 1),
        encoding="utf-8",
    )

    with pytest.raises(ObservabilityTomlError, match="missing key schema_version"):
        load_observability_contract(paths)


def test_rejects_unknown_filename(tmp_path: Path) -> None:
    paths = list(copied_observability_paths(tmp_path))
    unknown = tmp_path / "unknown.toml"
    unknown.write_text(paths[0].read_text(encoding="utf-8"), encoding="utf-8")
    paths[0] = unknown

    with pytest.raises(ObservabilityTomlError, match="filename must be one of"):
        load_observability_contract(paths)
