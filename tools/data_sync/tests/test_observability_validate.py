from __future__ import annotations

import shutil
from collections.abc import Callable
from pathlib import Path

import pytest

from data_sync.observability_validate import ObservabilityValidationError, validate_observability


OBSERVABILITY_PATHS = (
    "shared/contracts/observability/schema.toml",
    "shared/contracts/observability/services.toml",
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


def edit_source(paths: tuple[Path, ...], filename: str, old: str, new: str) -> None:
    path = next(path for path in paths if path.name == filename)
    text = path.read_text(encoding="utf-8")
    assert old in text, f"expected {old!r} in {filename}"
    path.write_text(text.replace(old, new, 1), encoding="utf-8")


def test_canonical_observability_validation_succeeds(tmp_path: Path) -> None:
    contract = validate_observability(copied_observability_paths(tmp_path))

    assert contract.schema.schema_version == 1
    assert len(contract.fields) == 47
    assert len(contract.services) == 5
    assert len(contract.events) == 151
    assert contract.event("diagnostic_report_stored").services == ("diagnostic_aggregator",)
    assert tuple(action.name for action in contract.redaction.actions) == ("reject", "redact")
    assert tuple(tier.name for tier in contract.retention_tiers) == (
        "ephemeral_dev",
        "operational",
        "diagnostic_report",
        "audit_grade",
    )


@pytest.mark.parametrize(
    ("filename", "old", "new", "message"),
    [
        (
            "fields.toml",
            '[[fields]]\nname = "level"',
            '[[fields]]\nname = "timestamp"',
            "duplicate field definition",
        ),
        (
            "events.toml",
            '[[events]]\nname = "service_started"',
            '[[events]]\nname = "service_starting"',
            "duplicate event",
        ),
        (
            "schema.toml",
            '  "build_version",\n',
            "",
            "required_fields does not exactly match",
        ),
        (
            "events.toml",
            'default_level = "info"',
            'default_level = "verbose"',
            "default level is not canonical",
        ),
        (
            "events.toml",
            'retention_tier = "operational"',
            'retention_tier = "missing_tier"',
            "retention tier is not declared",
        ),
        (
            "events.toml",
            'services = ["api_server", "game_server", "player_data", "diagnostic_aggregator"]',
            'services = ["unknown_service", "game_server", "player_data", "diagnostic_aggregator"]',
            "references unknown service keys",
        ),
        (
            "services.toml",
            'emitted_name = "player-data"',
            'emitted_name = "game-server"',
            "duplicate emitted service name",
        ),
        (
            "services.toml",
            'emitted_name = "game-server"',
            'emitted_name = "Game Server"',
            "emitted service name must be non-empty lowercase kebab-case",
        ),
        (
            "events.toml",
            "bridge_only = true",
            "bridge_only = false",
            "log_message must be explicitly declared bridge_only = true",
        ),
        (
            "events.toml",
            'services = ["client", "game_server", "player_data", "diagnostic_aggregator"]',
            'services = ["client", "game_server", "player_data"]',
            "log_message must be eligible for all legacy bridge components",
        ),
        (
            "schema.toml",
            '  "write_failed",\n',
            '  "unknown_event",\n',
            "duplicate rejection code",
        ),
        (
            "redaction.toml",
            'missing_action = "reject"',
            'missing_action = "missing_action"',
            "references undeclared redaction action",
        ),
        (
            "redaction.toml",
            "raw_forbidden_values_may_be_logged = false",
            "raw_forbidden_values_may_be_logged = true",
            "must not log rejected raw values",
        ),
        (
            "retention_tiers.toml",
            'default_age_seconds = 0\nmax_age_seconds = 0',
            'default_age_seconds = 10\nmax_age_seconds = 1',
            "default age must not exceed max age",
        ),
        (
            "retention_tiers.toml",
            "max_active_segment_age_seconds = 3600",
            "max_active_segment_age_seconds = 0",
            "file logging max active segment age must be positive",
        ),
        (
            "diagnostic_bundle.toml",
            'name = "timestamp"\ntype = "string"',
            'name = "timestamp"\ntype = "integer"',
            "type does not match canonical field",
        ),
        (
            "diagnostic_bundle.toml",
            'replacement_marker = "[REDACTED]"',
            'replacement_marker = "[REMOVED]"',
            "diagnostic replacement marker does not match redaction marker",
        ),
    ],
)
def test_semantic_validation_failure_is_clear(
    tmp_path: Path,
    filename: str,
    old: str,
    new: str,
    message: str,
) -> None:
    paths = copied_observability_paths(tmp_path)
    edit_source(paths, filename, old, new)

    with pytest.raises(ObservabilityValidationError) as exc_info:
        validate_observability(paths)

    assert exc_info.value.errors
    assert any(message in error for error in exc_info.value.errors)


def test_shared_operational_default_must_match_diagnostic_report(tmp_path: Path) -> None:
    paths = copied_observability_paths(tmp_path)
    edit_source(paths, "retention_tiers.toml", 'name = "operational"\npurpose = "Routine operational diagnosis and cross-service event investigation."\ndurability = "durable"\ncompression = "recommended"\ndelete_policy = "delete_when_configured_age_expires_or_storage_pressure_requires_eviction"\ndefault_age_seconds = 1209600', 'name = "operational"\npurpose = "Routine operational diagnosis and cross-service event investigation."\ndurability = "durable"\ncompression = "recommended"\ndelete_policy = "delete_when_configured_age_expires_or_storage_pressure_requires_eviction"\ndefault_age_seconds = 0')

    with pytest.raises(ObservabilityValidationError, match="operational default age must be 1209600"):
        validate_observability(paths)


def test_semantic_validation_accumulates_errors(tmp_path: Path) -> None:
    paths = copied_observability_paths(tmp_path)
    edit_source(paths, "fields.toml", '[[fields]]\nname = "level"', '[[fields]]\nname = "timestamp"')
    edit_source(paths, "events.toml", 'default_level = "info"', 'default_level = "verbose"')
    edit_source(
        paths,
        "retention_tiers.toml",
        'default_age_seconds = 0\nmax_age_seconds = 0',
        'default_age_seconds = 10\nmax_age_seconds = 1',
    )

    with pytest.raises(ObservabilityValidationError) as exc_info:
        validate_observability(paths)

    errors = "\n".join(exc_info.value.errors)
    assert "duplicate field definition" in errors
    assert "default level is not canonical" in errors
    assert "default age must not exceed max age" in errors
