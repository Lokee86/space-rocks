"""Rich packet TOML schema loading."""

from __future__ import annotations

from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Any
from typing import Iterable

from data_sync.model.packets import (
    PacketBuilder,
    PacketOutput,
    PacketSchema,
    PacketSchemaField,
    PacketStruct,
    PacketType,
)


class PacketTomlError(Exception):
    """Raised when packet TOML cannot be loaded."""


def load_packet_schema(path: Path | str) -> PacketSchema:
    tomlkit = _load_tomlkit()
    resolved_path = Path(path)
    document = _load_toml_document(tomlkit, resolved_path)

    return packet_schema_from_document(document)


def load_packet_schema_files(paths: Iterable[Path | str]) -> PacketSchema:
    resolved_paths = tuple(Path(path) for path in paths)
    if not resolved_paths:
        raise PacketTomlError("packet TOML paths must not be empty")

    tomlkit = _load_tomlkit()
    outputs: list[PacketOutput] = []
    structs: list[PacketStruct] = []
    packet_types: list[PacketType] = []
    builders: list[PacketBuilder] = []
    saw_rich_schema_list = False

    for resolved_path in resolved_paths:
        document = _load_toml_document(tomlkit, resolved_path)
        saw_rich_schema_list = saw_rich_schema_list or _contains_rich_schema_list(document)
        outputs.extend(
            _packet_output(item, index)
            for index, item in enumerate(_optional_list(document, "outputs"))
        )
        structs.extend(
            _packet_struct(item, index)
            for index, item in enumerate(_optional_list(document, "structs"))
        )
        packet_types.extend(
            _packet_type(item, index)
            for index, item in enumerate(_optional_list(document, "packet_types"))
        )
        builders.extend(
            _packet_builder(item, index)
            for index, item in enumerate(_optional_list(document, "builders"))
        )

    if not saw_rich_schema_list:
        raise PacketTomlError(
            "packet TOML does not contain rich schema lists: outputs, structs, packet_types, builders"
        )

    _validate_unique_output_ids(outputs)
    _validate_unique_output_paths(outputs)
    _validate_unique_values(
        (struct.id for struct in structs),
        "struct id",
    )
    _validate_unique_values(
        (packet_type.id for packet_type in packet_types),
        "packet_type id",
    )
    _validate_unique_values(
        (builder.id for builder in builders),
        "builder id",
    )

    return PacketSchema(
        outputs=tuple(outputs),
        structs=tuple(structs),
        packet_types=tuple(packet_types),
        builders=tuple(builders),
    )


def packet_schema_from_document(document: Mapping[str, Any]) -> PacketSchema:
    outputs = tuple(_packet_output(item, index) for index, item in enumerate(_required_list(document, "outputs")))
    structs = tuple(_packet_struct(item, index) for index, item in enumerate(_required_list(document, "structs")))
    packet_types = tuple(
        _packet_type(item, index) for index, item in enumerate(_required_list(document, "packet_types"))
    )
    builders = tuple(_packet_builder(item, index) for index, item in enumerate(_required_list(document, "builders")))
    return PacketSchema(
        outputs=outputs,
        structs=structs,
        packet_types=packet_types,
        builders=builders,
    )


def _load_toml_document(tomlkit: Any, resolved_path: Path) -> Mapping[str, Any]:
    try:
        text = resolved_path.read_text(encoding="utf-8")
    except FileNotFoundError as exc:
        raise PacketTomlError(f"packet TOML file does not exist: {resolved_path}") from exc
    except OSError as exc:
        raise PacketTomlError(f"failed to read packet TOML {resolved_path}: {exc}") from exc

    try:
        return tomlkit.parse(text)
    except Exception as exc:
        raise PacketTomlError(f"failed to parse packet TOML {resolved_path}: {exc}") from exc


def _contains_rich_schema_list(document: Mapping[str, Any]) -> bool:
    return any(
        key in document
        for key in ("outputs", "structs", "packet_types", "builders")
    )


def _packet_output(raw: Any, index: int) -> PacketOutput:
    table = _required_mapping(raw, f"outputs[{index}]")
    output_id = _optional_string(table.get("id"), f"outputs[{index}].id")
    language = _required_string(table, "language", f"outputs[{index}]")
    path = _required_string(table, "path", f"outputs[{index}]")
    imports = _optional_string_mapping(table.get("imports"), f"outputs[{index}].imports")
    extras = _extras(
        table,
        {
            "id",
            "language",
            "path",
            "package",
            "imports",
            "packet_types",
            "packet_type_ids",
            "structs",
            "base",
            "builders",
        },
    )

    return PacketOutput(
        id=output_id,
        language=language,
        path=path,
        package=_optional_string(table.get("package"), f"outputs[{index}].package"),
        imports=imports,
        packet_types=_optional_bool(table.get("packet_types"), f"outputs[{index}].packet_types", default=False),
        packet_type_ids=_string_tuple(table.get("packet_type_ids", []), f"outputs[{index}].packet_type_ids"),
        structs=_string_tuple(table.get("structs", []), f"outputs[{index}].structs"),
        base=_optional_string(table.get("base"), f"outputs[{index}].base"),
        builders=_string_tuple(table.get("builders", []), f"outputs[{index}].builders"),
        extras=extras,
    )


def _packet_struct(raw: Any, index: int) -> PacketStruct:
    table = _required_mapping(raw, f"structs[{index}]")
    struct_id = _required_string(table, "id", f"structs[{index}]")
    fields = tuple(
        _packet_field(item, struct_id, field_index)
        for field_index, item in enumerate(_required_list(table, "fields", f"structs[{index}]"))
    )
    return PacketStruct(
        id=struct_id,
        fields=fields,
        extras=_extras(table, {"id", "fields"}),
    )


def _packet_field(raw: Any, struct_id: str, index: int) -> PacketSchemaField:
    label = f"structs.{struct_id}.fields[{index}]"
    table = _required_mapping(raw, label)
    field_type = _required_string(table, "type", label)
    return PacketSchemaField(
        name=_required_string(table, "name", label),
        json=_required_string(table, "json", label),
        type=field_type,
        go_name=_optional_string(table.get("go_name"), f"{label}.go_name"),
        go_type=_optional_string(table.get("go_type"), f"{label}.go_type"),
        key_type=_optional_string(table.get("key_type"), f"{label}.key_type"),
        value_type=_optional_string(table.get("value_type"), f"{label}.value_type"),
        item_type=_optional_string(table.get("item_type"), f"{label}.item_type"),
        go_item_type=_optional_string(table.get("go_item_type"), f"{label}.go_item_type"),
        go_value_type=_optional_string(table.get("go_value_type"), f"{label}.go_value_type"),
        extras=_extras(
            table,
            {
                "name",
                "json",
                "type",
                "go_name",
                "go_type",
                "key_type",
                "value_type",
                "item_type",
                "go_item_type",
                "go_value_type",
            },
        ),
    )


def _packet_type(raw: Any, index: int) -> PacketType:
    table = _required_mapping(raw, f"packet_types[{index}]")
    return PacketType(
        id=_required_string(table, "id", f"packet_types[{index}]"),
        value=_required_string(table, "value", f"packet_types[{index}]"),
        extras=_extras(table, {"id", "value"}),
    )


def _packet_builder(raw: Any, index: int) -> PacketBuilder:
    table = _required_mapping(raw, f"builders[{index}]")
    builder_id = _required_string(table, "id", f"builders[{index}]")
    body = _required_mapping(table.get("body"), f"builders[{index}].body")
    return PacketBuilder(
        id=builder_id,
        args=_string_tuple(table.get("args", []), f"builders[{index}].args"),
        body=_plain_data(body),
        extras=_extras(table, {"id", "args", "body"}),
    )


def _required_list(table: Mapping[str, Any], key: str, label: str | None = None) -> Sequence[Any]:
    context = label or "packet TOML"
    value = table.get(key)
    if not isinstance(value, Sequence) or isinstance(value, (str, bytes, bytearray)):
        raise PacketTomlError(f"{context} missing list: {key}")
    return value


def _optional_list(table: Mapping[str, Any], key: str) -> Sequence[Any]:
    value = table.get(key)
    if value is None:
        return ()
    if not isinstance(value, Sequence) or isinstance(value, (str, bytes, bytearray)):
        raise PacketTomlError(f"packet TOML {key} must be a list")
    return value


def _validate_unique_output_ids(outputs: Sequence[PacketOutput]) -> None:
    _validate_unique_values(
        (output_id for output_id in (output.id for output in outputs) if output_id),
        "output id",
    )


def _validate_unique_output_paths(outputs: Sequence[PacketOutput]) -> None:
    _validate_unique_values((output.path for output in outputs), "output path")


def _validate_unique_values(values: Iterable[str], kind: str) -> None:
    seen: set[str] = set()
    for value in values:
        if value in seen:
            raise PacketTomlError(f"duplicate {kind}: {value}")
        seen.add(value)


def _required_mapping(value: Any, label: str) -> Mapping[str, Any]:
    if not _is_mapping(value):
        raise PacketTomlError(f"{label} must be a table")
    return value


def _required_string(table: Mapping[str, Any], key: str, label: str) -> str:
    value = _plain_data(table.get(key))
    if not isinstance(value, str) or not value:
        raise PacketTomlError(f"{label}.{key} must be a non-empty string")
    return value


def _optional_string(value: Any, label: str) -> str | None:
    if value is None:
        return None
    value = _plain_data(value)
    if not isinstance(value, str) or not value:
        raise PacketTomlError(f"{label} must be a non-empty string")
    return value


def _optional_bool(value: Any, label: str, *, default: bool) -> bool:
    if value is None:
        return default
    value = _plain_data(value)
    if not isinstance(value, bool):
        raise PacketTomlError(f"{label} must be a boolean")
    return value


def _optional_string_mapping(value: Any, label: str) -> Mapping[str, str] | None:
    if value is None:
        return None
    table = _required_mapping(value, label)
    result: dict[str, str] = {}
    for key, item in table.items():
        item = _plain_data(item)
        if not isinstance(item, str) or not item:
            raise PacketTomlError(f"{label}.{key} must be a non-empty string")
        result[key] = item
    return result


def _string_tuple(value: Any, label: str) -> tuple[str, ...]:
    if not isinstance(value, Sequence) or isinstance(value, (str, bytes, bytearray)):
        raise PacketTomlError(f"{label} must be a list of strings")
    result = tuple(_plain_data(item) for item in value)
    if not all(isinstance(item, str) and item for item in result):
        raise PacketTomlError(f"{label} must contain only non-empty strings")
    return result


def _extras(table: Mapping[str, Any], known_keys: set[str]) -> Mapping[str, Any] | None:
    extras = {key: _plain_data(value) for key, value in table.items() if key not in known_keys}
    return extras or None


def _plain_data(value: Any) -> Any:
    if hasattr(value, "unwrap"):
        value = value.unwrap()
    if _is_mapping(value):
        return {key: _plain_data(item) for key, item in value.items()}
    if isinstance(value, Sequence) and not isinstance(value, (str, bytes, bytearray)):
        return [_plain_data(item) for item in value]
    return value


def _is_mapping(value: Any) -> bool:
    return hasattr(value, "items") and hasattr(value, "__contains__")


def _load_tomlkit() -> Any:
    try:
        import tomlkit

        return tomlkit
    except ModuleNotFoundError as exc:
        raise PacketTomlError("tomlkit is required for reading packet TOML") from exc
