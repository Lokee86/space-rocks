"""Validation command support."""

from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path

from data_sync.block_io import BlockIOError, find_block
from data_sync.cli import DOMAINS
from data_sync.config import DataSyncConfig
from data_sync.model.constants import ConstantValue
from data_sync.model.packets import PacketDefinition, PacketSchema, PacketSchemaField
from data_sync.packet_rendering import GO_PRIMITIVES, PacketRenderingError, parse_rich_type
from data_sync.packet_toml import PacketTomlError, load_packet_schema
from data_sync.toml_store import TomlStore, TomlStoreError


SNAKE_CASE_RE = re.compile(r"^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$")
CONSTANT_VALUE_TYPES = (int, float, bool, str)
PACKET_DIRECTIONS = {"client_to_server", "server_to_client", "bidirectional"}
PACKET_FIELD_TYPES = {"bool", "int", "uint32", "float32", "float64", "string"}
SUPPORTED_PACKET_LANGUAGES = {"go", "gdscript"}
STRUCT_NAME_RE = re.compile(r"^[A-Z][A-Za-z0-9]*$")
GO_IMPORT_ALIAS_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")


class ValidationError(Exception):
    """Raised when validation fails."""

    def __init__(self, errors: list[str]) -> None:
        self.errors = errors
        super().__init__("\n".join(errors))


@dataclass(frozen=True)
class ValidationRequest:
    domains: tuple[str, ...]
    languages: tuple[str, ...]


def validate(config: DataSyncConfig, domains: tuple[str, ...], languages: tuple[str, ...]) -> None:
    requested_domains = domains or _enabled_domains(config)
    request = ValidationRequest(
        domains=requested_domains,
        languages=languages,
    )
    errors: list[str] = []

    if "constants" in request.domains:
        constants_store = _load_store(config.sot_path("constants"), errors)
        if constants_store is not None:
            _validate_constants(config, constants_store, request, errors)
    if "packets" in request.domains:
        _validate_packet_sot(config.sot_path("packets"), errors)

    _validate_configured_files_and_blocks(config, request, errors)

    if errors:
        raise ValidationError(errors)


def _load_store(path: Path, errors: list[str]) -> TomlStore | None:
    try:
        return TomlStore.load(path)
    except TomlStoreError as exc:
        errors.append(str(exc))
        return None


def _validate_packet_sot(path: Path, errors: list[str]) -> None:
    try:
        schema = load_packet_schema(path)
        _validate_rich_packet_schema(schema, errors)
        return
    except PacketTomlError as exc:
        rich_error = str(exc)

    packets_store = _load_store(path, errors)
    if packets_store is not None:
        try:
            if packets_store.packets():
                _validate_packets(packets_store, errors)
                return
        except TomlStoreError:
            pass
    errors.append(rich_error)


def _validate_rich_packet_schema(schema: PacketSchema, errors: list[str]) -> None:
    struct_ids = {struct.id for struct in schema.structs}
    packet_type_ids: dict[str, str] = {}
    packet_type_values: dict[str, str] = {}
    builder_ids = {builder.id for builder in schema.builders}
    json_fields = {field.json: field for struct in schema.structs for field in struct.fields}
    fields_by_struct = {
        struct.id: {field.json: field for field in struct.fields}
        for struct in schema.structs
    }

    if not schema.outputs:
        errors.append("packet TOML must contain at least one output")
    if not schema.structs:
        errors.append("packet TOML must contain at least one struct")
    if not schema.packet_types:
        errors.append("packet TOML must contain at least one packet_type")

    for output in schema.outputs:
        _validate_packet_output(output, struct_ids, builder_ids, errors)

    for struct in schema.structs:
        if not STRUCT_NAME_RE.fullmatch(struct.id):
            errors.append(f"packet struct name is invalid: {struct.id}")
        if not struct.fields:
            errors.append(f"packet struct {struct.id} must contain at least one field")
        for field in struct.fields:
            _validate_schema_field(struct.id, field, struct_ids, errors)

    for packet_type in schema.packet_types:
        if not _is_snake_case(packet_type.id):
            errors.append(f"packet type name is not valid snake_case: {packet_type.id}")
        previous_id = packet_type_ids.get(packet_type.id)
        if previous_id is not None:
            errors.append(f"duplicate packet type id: {packet_type.id}")
        packet_type_ids[packet_type.id] = packet_type.value
        previous_value = packet_type_values.get(packet_type.value)
        if previous_value is not None:
            errors.append(
                f"duplicate packet type value {packet_type.value!r}: {previous_value}, {packet_type.id}"
            )
        packet_type_values[packet_type.value] = packet_type.id

    for builder in schema.builders:
        if not builder.id or not builder.id.endswith("_packet") or not _is_snake_case(builder.id):
            errors.append(f"packet builder name is invalid: {builder.id}")
        _validate_builder_body(
            f"builder {builder.id}",
            builder.body,
            set(builder.args),
            set(packet_type_values),
            json_fields,
            fields_by_struct,
            None,
            errors,
        )


def _validate_packet_output(
    output,
    struct_ids: set[str],
    builder_ids: set[str],
    errors: list[str],
) -> None:
    if output.language not in SUPPORTED_PACKET_LANGUAGES:
        errors.append(f"unsupported packet output language: {output.language}")
    if not output.path or Path(output.path).is_absolute() or ".." in Path(output.path).parts:
        errors.append(f"packet output path must be a relative project path: {output.path}")
    if output.language == "go":
        if not output.package:
            errors.append(f"Go packet output missing package: {output.path}")
        if output.package and not GO_IMPORT_ALIAS_RE.fullmatch(output.package):
            errors.append(f"Go packet output package is invalid: {output.package}")
        for alias, import_path in (output.imports or {}).items():
            if not GO_IMPORT_ALIAS_RE.fullmatch(alias):
                errors.append(f"Go packet import alias is invalid: {alias}")
            if not import_path:
                errors.append(f"Go packet import path for {alias} must be non-empty")
        for struct_id in output.structs:
            if struct_id not in struct_ids:
                errors.append(f"packet output {output.path} references unknown struct: {struct_id}")
    if output.language == "gdscript":
        for builder_id in output.builders:
            if builder_id not in builder_ids:
                errors.append(f"packet output {output.path} references unknown builder: {builder_id}")


def _validate_schema_field(
    struct_id: str,
    field: PacketSchemaField,
    struct_ids: set[str],
    errors: list[str],
) -> None:
    if not _is_snake_case(field.name):
        errors.append(f"{struct_id}.{field.name} is not a valid snake_case field name")
    if not _is_json_field_name(field.json):
        errors.append(f"{struct_id}.{field.name} has invalid JSON field name: {field.json}")
    _validate_field_type(struct_id, field, struct_ids, errors)


def _validate_field_type(
    struct_id: str,
    field: PacketSchemaField,
    struct_ids: set[str],
    errors: list[str],
) -> None:
    if field.type in {"map", "dictionary"}:
        if not field.key_type or not field.value_type:
            errors.append(f"{struct_id}.{field.name} map field requires key_type and value_type")
            return
        _validate_scalar_type(struct_id, field.name, field.key_type, "key_type", errors)
        _validate_type_ref(struct_id, field.name, field.value_type, struct_ids, errors)
        return
    if field.type in {"array", "list"}:
        if not field.item_type:
            errors.append(f"{struct_id}.{field.name} array field requires item_type")
            return
        _validate_type_ref(struct_id, field.name, field.item_type, struct_ids, errors)
        return

    try:
        rich_type = parse_rich_type(field.type)
    except PacketRenderingError as exc:
        errors.append(f"{struct_id}.{field.name} has invalid rich type string: {exc}")
        return
    if rich_type is not None:
        kind, args = rich_type
        if kind in {"map", "dictionary"}:
            if len(args) != 2:
                errors.append(f"{struct_id}.{field.name} map type requires key and value types")
                return
            _validate_scalar_type(struct_id, field.name, args[0], "key type", errors)
            _validate_type_ref(struct_id, field.name, args[1], struct_ids, errors)
            return
        if kind in {"array", "list"}:
            if len(args) != 1:
                errors.append(f"{struct_id}.{field.name} array type requires one item type")
                return
            _validate_type_ref(struct_id, field.name, args[0], struct_ids, errors)
            return
        errors.append(f"{struct_id}.{field.name} has unsupported rich type kind: {kind}")
        return

    _validate_type_ref(struct_id, field.name, field.type, struct_ids, errors)


def _validate_scalar_type(
    struct_id: str,
    field_name: str,
    value: str,
    label: str,
    errors: list[str],
) -> None:
    if value not in GO_PRIMITIVES:
        errors.append(f"{struct_id}.{field_name} has unsupported {label}: {value}")


def _validate_type_ref(
    struct_id: str,
    field_name: str,
    value: str,
    struct_ids: set[str],
    errors: list[str],
) -> None:
    if value in GO_PRIMITIVES:
        return
    if value in struct_ids:
        return
    errors.append(f"{struct_id}.{field_name} references unknown packet type/struct: {value}")


def _validate_builder_body(
    label: str,
    body,
    args: set[str],
    packet_values: set[str],
    json_fields: dict[str, PacketSchemaField],
    fields_by_struct: dict[str, dict[str, PacketSchemaField]],
    current_struct: str | None,
    errors: list[str],
) -> None:
    if not hasattr(body, "items"):
        errors.append(f"{label} body must be a dictionary")
        return
    allowed_fields = fields_by_struct.get(current_struct, json_fields)
    for key, value in body.items():
        if key not in allowed_fields:
            errors.append(f"{label} references unknown packet field: {key}")
            continue
        field = allowed_fields[key]
        if hasattr(value, "items"):
            nested_struct = field.type if field.type in fields_by_struct else None
            _validate_builder_body(
                f"{label}.{key}",
                value,
                args,
                packet_values,
                json_fields,
                fields_by_struct,
                nested_struct,
                errors,
            )
            continue
        if isinstance(value, str) and value.startswith("$"):
            arg_name = value[1:]
            if arg_name not in args:
                errors.append(f"{label}.{key} references unknown builder arg: {arg_name}")
        elif key == "type" and isinstance(value, str) and value not in packet_values:
            errors.append(f"{label}.{key} references unknown packet type value: {value}")


def _validate_constants(
    config: DataSyncConfig,
    store: TomlStore,
    request: ValidationRequest,
    errors: list[str],
) -> None:
    section_names = _requested_sections(
        config,
        "constants",
        _languages_for_domain(config, "constants", request.languages),
    )
    for section_name in section_names:
        try:
            section = store.constants(section_name)
        except TomlStoreError as exc:
            errors.append(str(exc))
            continue

        if not section.values:
            errors.append(f"[{section_name}] must contain at least one constant")
        for name, value in section.values:
            if not _is_snake_case(name):
                errors.append(f"[{section_name}].{name} is not a valid snake_case constant name")
            if not _is_supported_constant_value(value):
                errors.append(
                    f"[{section_name}].{name} has unsupported value type: {type(value).__name__}"
                )


def _validate_packets(store: TomlStore, errors: list[str]) -> None:
    try:
        packets = store.packets()
    except TomlStoreError as exc:
        errors.append(str(exc))
        return

    seen_ids: dict[int | str, str] = {}
    for packet in packets:
        _validate_packet(packet, seen_ids, errors)


def _validate_packet(
    packet: PacketDefinition,
    seen_ids: dict[int | str, str],
    errors: list[str],
) -> None:
    if not _is_snake_case(packet.name):
        errors.append(f"[packets.{packet.name}] is not a valid snake_case packet name")

    previous = seen_ids.get(packet.id)
    if previous is not None:
        errors.append(f"duplicate packet id {packet.id!r}: packets.{previous}, packets.{packet.name}")
    else:
        seen_ids[packet.id] = packet.name

    if packet.direction not in PACKET_DIRECTIONS:
        errors.append(
            f"[packets.{packet.name}].direction must be one of: {', '.join(sorted(PACKET_DIRECTIONS))}"
        )

    for field in packet.fields:
        if not _is_snake_case(field.name):
            errors.append(
                f"[packets.{packet.name}.fields].{field.name} is not a valid snake_case field name"
            )
        if field.type not in PACKET_FIELD_TYPES:
            errors.append(
                f"[packets.{packet.name}.fields].{field.name} has unsupported field type: {field.type}"
            )


def _validate_configured_files_and_blocks(
    config: DataSyncConfig,
    request: ValidationRequest,
    errors: list[str],
) -> None:
    for domain in request.domains:
        for language in _languages_for_domain(config, domain, request.languages):
            target = config.target(domain, language)
            for path in target.files:
                text = _read_configured_file(path, errors)
                if text is None:
                    continue
                if domain == "packets":
                    continue
                for section_name in target.sections:
                    try:
                        find_block(text, section_name)
                    except BlockIOError as exc:
                        errors.append(f"{path}: {exc}")


def _requested_sections(
    config: DataSyncConfig,
    domain: str,
    languages: tuple[str, ...],
) -> tuple[str, ...]:
    seen: set[str] = set()
    sections: list[str] = []
    for language in languages:
        target = config.target(domain, language)
        for section_name in target.sections:
            if section_name not in seen:
                seen.add(section_name)
                sections.append(section_name)
    return tuple(sections)


def _languages_for_domain(
    config: DataSyncConfig,
    domain: str,
    requested_languages: tuple[str, ...],
) -> tuple[str, ...]:
    languages = requested_languages or config.enabled_languages(domain)
    return tuple(language for language in languages if config.target(domain, language).enabled)


def _enabled_domains(config: DataSyncConfig) -> tuple[str, ...]:
    return tuple(domain for domain in DOMAINS if config.enabled_languages(domain))


def _read_configured_file(path: Path, errors: list[str]) -> str | None:
    try:
        return path.read_text(encoding="utf-8")
    except FileNotFoundError:
        errors.append(f"configured file does not exist: {path}")
        return None
    except OSError as exc:
        errors.append(f"failed to read configured file {path}: {exc}")
        return None


def _is_snake_case(value: str) -> bool:
    return bool(SNAKE_CASE_RE.fullmatch(value))


def _is_json_field_name(value: str) -> bool:
    return bool(SNAKE_CASE_RE.fullmatch(value))


def _is_supported_constant_value(value: ConstantValue) -> bool:
    if isinstance(value, bool):
        return True
    if isinstance(value, (int, float, str)):
        return True
    if (
        isinstance(value, list)
        and len(value) == 2
        and all(isinstance(item, (int, float)) and not isinstance(item, bool) for item in value)
    ):
        return True
    return False
