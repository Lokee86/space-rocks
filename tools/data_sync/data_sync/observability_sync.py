from __future__ import annotations

from collections.abc import Callable

from data_sync.config import ConfigError, DataSyncConfig
from data_sync.constants_sync import FileUpdate
from data_sync.generators.observability_docs import generate_observability_docs
from data_sync.generators.observability_gds import generate_observability_gds
from data_sync.generators.observability_go import generate_observability_go
from data_sync.generators.observability_json import generate_observability_json
from data_sync.generators.observability_ruby import generate_observability_ruby
from data_sync.model.observability import ObservabilityContract
from data_sync.observability_toml import ObservabilityTomlError
from data_sync.observability_validate import ObservabilityValidationError, validate_observability


class ObservabilitySyncError(Exception):
    pass


_RENDERERS: dict[str, Callable[[ObservabilityContract], str]] = {
    "go": generate_observability_go,
    "gds": generate_observability_gds,
    "ruby": generate_observability_ruby,
    "json": generate_observability_json,
    "docs": generate_observability_docs,
}
_OUTPUT_ORDER = {name: index for index, name in enumerate(("go", "gds", "ruby", "json", "docs"))}


def plan_observability_updates(
    config: DataSyncConfig, outputs: tuple[str, ...]
) -> tuple[FileUpdate, ...]:
    try:
        contract = validate_observability(config.sot_paths("observability"))
    except (ConfigError, ObservabilityTomlError, ObservabilityValidationError) as exc:
        raise ObservabilitySyncError(str(exc)) from exc

    updates: list[FileUpdate] = []
    generated_by_kind: dict[str, str] = {}
    for output in sorted(outputs, key=lambda item: _OUTPUT_ORDER.get(item, len(_OUTPUT_ORDER))):
        renderer = _RENDERERS.get(output)
        if renderer is None:
            raise ObservabilitySyncError(f"unsupported observability output kind: {output}")
        if output not in generated_by_kind:
            generated_by_kind[output] = renderer(contract)
        try:
            targets = config.targets_for("observability", output)
        except ConfigError as exc:
            raise ObservabilitySyncError(str(exc)) from exc
        for target in sorted(targets, key=lambda item: item.label):
            if not target.enabled:
                raise ObservabilitySyncError(f"[{target.label}] is disabled in config")
            for path in sorted(target.files):
                before = path.read_text(encoding="utf-8") if path.exists() else ""
                updates.append(
                    FileUpdate(path=path, before=before, after=generated_by_kind[output])
                )
    return tuple(updates)
