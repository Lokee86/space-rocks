"""Player-data TOML loading."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any, Mapping
import tomllib


SUPPORTED_FIELD_TYPES = {"string", "integer", "boolean"}


class PlayerDataTomlError(Exception):
    """Raised when player-data TOML cannot be loaded."""


@dataclass(frozen=True)
class PlayerDataField:
    name: str
    type: str
    required: bool = False
    default: Any | None = None


@dataclass(frozen=True)
class PlayerDataGroup:
    name: str
    fields: tuple[PlayerDataField, ...]


@dataclass(frozen=True)
class PlayerDataSchema:
    name: str
    version: str | None
    groups: tuple[PlayerDataGroup, ...]


def load_player_data_schema(path: Path | str) -> PlayerDataSchema:
    resolved_path = Path(path)
    document = _load_toml_document(resolved_path)
    schema_name = _required_schema_name(document)
    schema_version = _optional_string(document.get("schema_version"), "schema_version")
    groups = _load_groups(document, schema_name)
    return PlayerDataSchema(name=schema_name, version=schema_version, groups=groups)


def _load_toml_document(resolved_path: Path) -> Mapping[str, Any]:
    try:
        text = resolved_path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise PlayerDataTomlError(f"player-data TOML file does not exist: {resolved_path}") from exc
    except OSError as exc:
        raise PlayerDataTomlError(f"failed to read player-data TOML {resolved_path}: {exc}") from exc

    try:
        return tomllib.loads(text)
    except tomllib.TOMLDecodeError as exc:
        raise PlayerDataTomlError(f"failed to parse player-data TOML {resolved_path}: {exc}") from exc


def _required_schema_name(document: Mapping[str, Any]) -> str:
    for key in ("schema_name", "schema_id", "name"):
        if key in document:
            return _required_string(document.get(key), key)
    raise PlayerDataTomlError("player-data TOML must define schema_name, schema_id, or name")


def _load_groups(document: Mapping[str, Any], schema_name: str) -> tuple[PlayerDataGroup, ...]:
    groups: list[PlayerDataGroup] = []

    if "fields" in document:
        groups.append(PlayerDataGroup(name=schema_name, fields=_load_fields(document["fields"], "[fields]")))

    for key, value in document.items():
        if key in {"schema_name", "schema_id", "name", "schema_version", "fields"}:
            continue
        if not _is_mapping(value):
            continue

        fields_table = value.get("fields")
        if fields_table is None:
            continue
        groups.append(PlayerDataGroup(name=key, fields=_load_fields(fields_table, f"[{key}.fields]")))

    if not groups:
        raise PlayerDataTomlError("player-data TOML must define at least one field group")

    return tuple(groups)


def _load_fields(raw_fields: Any, label: str) -> tuple[PlayerDataField, ...]:
    fields_table = _required_mapping(raw_fields, label)
    fields: list[PlayerDataField] = []

    for raw_name, raw_field in fields_table.items():
        field_name = _required_field_name(raw_name, label)
        field_table = _required_mapping(raw_field, f"{label}.{field_name}")
        field_type = _required_field_type(field_table.get("type"), f"{label}.{field_name}.type")
        _validate_supported_type(field_type, f"{label}.{field_name}.type")

        has_required = "required" in field_table
        has_default = "default" in field_table
        has_optional = "optional" in field_table
        if not has_required and not has_default and not has_optional:
            raise PlayerDataTomlError(
                f"{label}.{field_name} must define required, default, or optional"
            )

        required = False
        if has_required:
            required = _required_bool(field_table.get("required"), f"{label}.{field_name}.required")
        elif has_optional:
            optional = _required_bool(field_table.get("optional"), f"{label}.{field_name}.optional")
            required = not optional

        default = field_table.get("default") if has_default else None
        if has_default:
            _validate_default_type(default, field_type, f"{label}.{field_name}.default")

        fields.append(
            PlayerDataField(
                name=field_name,
                type=field_type,
                required=required,
                default=default,
            )
        )

    if not fields:
        raise PlayerDataTomlError(f"{label} must define at least one field")

    return tuple(fields)


def _required_field_name(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value:
        raise PlayerDataTomlError(f"{label} contains a field with a missing name")
    return value


def _required_field_type(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value:
        raise PlayerDataTomlError(f"{label} must be a non-empty string")
    return value


def _validate_supported_type(field_type: str, label: str) -> None:
    if field_type not in SUPPORTED_FIELD_TYPES:
        raise PlayerDataTomlError(f"{label} must be one of: string, integer, boolean")


def _validate_default_type(value: Any, field_type: str, label: str) -> None:
    if field_type == "string":
        if not isinstance(value, str):
            raise PlayerDataTomlError(f"{label} must be a string")
        return
    if field_type == "integer":
        if not isinstance(value, int) or isinstance(value, bool):
            raise PlayerDataTomlError(f"{label} must be an integer")
        return
    if field_type == "boolean":
        if not isinstance(value, bool):
            raise PlayerDataTomlError(f"{label} must be a boolean")
        return


def _required_mapping(value: Any, label: str) -> Mapping[str, Any]:
    if not _is_mapping(value):
        raise PlayerDataTomlError(f"{label} must be a table")
    return value


def _required_string(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value:
        raise PlayerDataTomlError(f"{label} must be a non-empty string")
    return value


def _required_bool(value: Any, label: str) -> bool:
    if not isinstance(value, bool):
        raise PlayerDataTomlError(f"{label} must be a boolean")
    return value


def _optional_string(value: Any, label: str) -> str | None:
    if value is None:
        return None
    return _required_string(value, label)


def _is_mapping(value: Any) -> bool:
    return isinstance(value, Mapping)
