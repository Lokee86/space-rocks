from __future__ import annotations

import shutil
from pathlib import Path

import pytest

from data_sync.config import DEFAULT_CONSTANTS_SCAN, DataSyncConfig, DomainLanguageConfig
from data_sync.constants_sync import FileUpdate, apply_updates
from data_sync.observability_sync import ObservabilitySyncError, plan_observability_updates


REPO = Path(__file__).resolve().parents[3]
OBSERVABILITY_SOURCE_NAMES = (
    "schema.toml",
    "services.toml",
    "events.toml",
    "fields.toml",
    "redaction.toml",
    "retention_tiers.toml",
    "diagnostic_bundle.toml",
)
OUTPUTS = ("go", "gds", "ruby", "json", "docs")


def _temporary_config(tmp_path: Path) -> DataSyncConfig:
    source_dir = tmp_path / "shared" / "contracts" / "observability"
    source_dir.mkdir(parents=True)
    source_paths = []
    for name in OBSERVABILITY_SOURCE_NAMES:
        source = REPO / "shared" / "contracts" / "observability" / name
        destination = source_dir / name
        shutil.copyfile(source, destination)
        source_paths.append(destination)

    target_files = {
        "go": (
            tmp_path / "generated" / "go" / "contract.go",
        ),
        "gds": (tmp_path / "generated" / "gds" / "contract.gd",),
        "ruby": (tmp_path / "generated" / "ruby" / "contract.rb",),
        "json": (tmp_path / "generated" / "json" / "contract.json",),
        "docs": (tmp_path / "generated" / "docs" / "contract.md",),
    }
    targets = {
        ("observability", output): (
            DomainLanguageConfig(
                domain="observability",
                language=output,
                label=f"observability.{output}",
                files=files,
                sections=(),
                owns=(),
            ),
        )
        for output, files in target_files.items()
    }
    return DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={"observability": tuple(source_paths)},
        targets_by_domain_language=targets,
        constants_scan=DEFAULT_CONSTANTS_SCAN,
    )


def test_observability_sync_plans_applies_and_detects_drift(tmp_path: Path) -> None:
    config = _temporary_config(tmp_path)

    first_plan = plan_observability_updates(config, OUTPUTS)
    second_plan = plan_observability_updates(config, OUTPUTS)

    assert len(first_plan) == 5
    assert all(isinstance(update, FileUpdate) for update in first_plan)
    assert first_plan == second_plan
    assert all(update.before == "" for update in first_plan)
    go_updates = first_plan[:1]
    assert len({update.after for update in go_updates}) == 1

    apply_updates(first_plan)
    synced_plan = plan_observability_updates(config, OUTPUTS)

    for update in synced_plan:
        assert "diagnostic_report_stored" in update.after
    assert len(synced_plan) == 5
    assert all(not update.changed for update in synced_plan)
    assert all(update.path.is_file() for update in synced_plan)

    events_path = config.sot_paths("observability")[2]
    events = events_path.read_text(encoding="utf-8")
    events_path.write_text(events.replace("service_starting", "service_starting_changed", 1), encoding="utf-8")

    drift_plan = plan_observability_updates(config, OUTPUTS)
    assert len(drift_plan) == 5
    assert all(update.changed for update in drift_plan)


def test_observability_sync_rejects_unsupported_output(tmp_path: Path) -> None:
    config = _temporary_config(tmp_path)

    with pytest.raises(ObservabilitySyncError, match="unsupported observability output kind"):
        plan_observability_updates(config, ("yaml",))


def test_observability_sync_rejects_disabled_target(tmp_path: Path) -> None:
    config = _temporary_config(tmp_path)
    target = config.targets_by_domain_language[("observability", "json")][0]
    disabled = DomainLanguageConfig(
        domain=target.domain,
        language=target.language,
        label=target.label,
        files=target.files,
        sections=target.sections,
        owns=target.owns,
        enabled=False,
    )
    targets = dict(config.targets_by_domain_language)
    targets[("observability", "json")] = (disabled,)
    disabled_config = DataSyncConfig(
        path=config.path,
        root=config.root,
        sot_paths_by_domain=config.sot_paths_by_domain,
        targets_by_domain_language=targets,
        constants_scan=config.constants_scan,
    )

    with pytest.raises(ObservabilitySyncError, match=r"\[observability\.json\] is disabled"):
        plan_observability_updates(disabled_config, ("json",))
