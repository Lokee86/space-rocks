"""Drop table TOML loading."""

from __future__ import annotations

from pathlib import Path
from typing import Any

from data_sync.model.drop_tables import DropTable, DropTableEntry, DropTablesModel


class DropTablesTomlError(Exception):
    """Raised when drop table TOML cannot be loaded."""


def load_drop_table(path: Path | str) -> DropTable:
    tomlkit = _load_tomlkit()
    resolved_path = Path(path)
    document = _load_toml_document(tomlkit, resolved_path)
    table = _required_mapping(document.get("table"), "[table]")

    table_id = _required_string(table.get("id"), "[table].id")
    source_type = _required_string(table.get("source_type"), "[table].source_type")
    drop_mode = _required_drop_mode(table.get("drop_mode"), "[table].drop_mode")
    max_drops_per_source = _required_int(
        table.get("max_drops_per_source"),
        "[table].max_drops_per_source",
    )
    _validate_min_drops_per_source(max_drops_per_source, "[table].max_drops_per_source")
    max_active_pickups = _required_int(table.get("max_active_pickups"), "[table].max_active_pickups")
    _validate_non_negative(max_active_pickups, "[table].max_active_pickups")

    entries_value = document.get("entries", ())
    entries_raw = _required_sequence(entries_value, "[[entries]]")
    entries = tuple(_entry_from_table(item, index) for index, item in enumerate(entries_raw))

    return DropTable(
        id=table_id,
        source_type=source_type,
        drop_mode=drop_mode,
        max_drops_per_source=max_drops_per_source,
        max_active_pickups=max_active_pickups,
        entries=entries,
    )


def load_drop_tables(paths: list[Path | str]) -> DropTablesModel:
    if not paths:
        raise DropTablesTomlError("drop table TOML paths must not be empty")

    tables: list[DropTable] = []
    seen_ids: set[str] = set()

    for path in paths:
        table = load_drop_table(path)
        if table.id in seen_ids:
            raise DropTablesTomlError(f"duplicate drop table id: {table.id}")
        seen_ids.add(table.id)
        tables.append(table)

    return DropTablesModel(tables=tuple(tables))


def _entry_from_table(raw: Any, index: int) -> DropTableEntry:
    label = f"[[entries]][{index}]"
    table = _required_mapping(raw, label)
    pickup_type = _required_string(table.get("pickup_type"), f"{label}.pickup_type")
    chance = _required_number(table.get("chance"), f"{label}.chance")
    min_source_size = _required_int(table.get("min_source_size"), f"{label}.min_source_size")
    max_source_size = _required_int(table.get("max_source_size"), f"{label}.max_source_size")
    _validate_chance(chance, f"{label}.chance")
    _validate_source_size_range(min_source_size, max_source_size, label)
    return DropTableEntry(
        pickup_type=pickup_type,
        chance=chance,
        min_source_size=min_source_size,
        max_source_size=max_source_size,
    )


def _load_toml_document(tomlkit: Any, resolved_path: Path) -> Any:
    try:
        text = resolved_path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise DropTablesTomlError(f"drop table TOML file does not exist: {resolved_path}") from exc
    except OSError as exc:
        raise DropTablesTomlError(f"failed to read drop table TOML {resolved_path}: {exc}") from exc

    try:
        return tomlkit.parse(text)
    except Exception as exc:
        raise DropTablesTomlError(f"failed to parse drop table TOML {resolved_path}: {exc}") from exc


def _required_mapping(value: Any, label: str) -> Any:
    if not _is_mapping(value):
        raise DropTablesTomlError(f"{label} must be a table")
    return value


def _required_sequence(value: Any, label: str) -> tuple[Any, ...]:
    if not isinstance(value, (list, tuple)):
        raise DropTablesTomlError(f"{label} must be a list")
    return tuple(value)


def _required_string(value: Any, label: str) -> str:
    value = _plain_data(value)
    if not isinstance(value, str) or not value:
        raise DropTablesTomlError(f"{label} must be a non-empty string")
    return value


def _required_int(value: Any, label: str) -> int:
    value = _plain_data(value)
    if not isinstance(value, int) or isinstance(value, bool):
        raise DropTablesTomlError(f"{label} must be an integer")
    return value


def _required_number(value: Any, label: str) -> float:
    value = _plain_data(value)
    if not isinstance(value, (int, float)) or isinstance(value, bool):
        raise DropTablesTomlError(f"{label} must be a number")
    return float(value)


def _validate_non_negative(value: int, label: str) -> None:
    if value < 0:
        raise DropTablesTomlError(f"{label} must be greater than or equal to 0")


def _required_drop_mode(value: Any, label: str) -> str:
    value = _required_string(value, label)
    if value not in {"single", "multi"}:
        raise DropTablesTomlError(f"{label} must be one of: single, multi")
    return value


def _validate_min_drops_per_source(value: int, label: str) -> None:
    if value < 1:
        raise DropTablesTomlError(f"{label} must be greater than or equal to 1")


def _validate_chance(value: float, label: str) -> None:
    if value < 0.0 or value > 1.0:
        raise DropTablesTomlError(f"{label} must be between 0.0 and 1.0")


def _validate_source_size_range(min_source_size: int, max_source_size: int, label: str) -> None:
    if min_source_size > max_source_size:
        raise DropTablesTomlError(
            f"{label}.min_source_size must be less than or equal to {label}.max_source_size"
        )


def _plain_data(value: Any) -> Any:
    if hasattr(value, "unwrap"):
        value = value.unwrap()
    return value


def _is_mapping(value: Any) -> bool:
    return hasattr(value, "items") and hasattr(value, "__contains__")


def _load_tomlkit() -> Any:
    try:
        import tomlkit

        return tomlkit
    except ModuleNotFoundError as exc:
        raise DropTablesTomlError("tomlkit is required for reading drop table TOML") from exc
