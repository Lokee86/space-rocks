"""Validation command support."""

from __future__ import annotations

import re
from dataclasses import dataclass
from pathlib import Path

from data_sync.block_io import BlockIOError, find_block
from data_sync.cli import DOMAINS
from data_sync.config import DataSyncConfig
from data_sync.model.constants import ConstantValue
from data_sync.model.packets import PacketDefinition
from data_sync.toml_store import TomlStore, TomlStoreError


SNAKE_CASE_RE = re.compile(r"^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$")
CONSTANT_VALUE_TYPES = (int, float, bool, str)
PACKET_DIRECTIONS = {"client_to_server", "server_to_client", "bidirectional"}
PACKET_FIELD_TYPES = {"bool", "int", "uint32", "float32", "float64", "string"}


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

    try:
        store = TomlStore.load(config.sot_path)
    except TomlStoreError as exc:
        raise ValidationError([str(exc)]) from exc

    if "constants" in request.domains:
        _validate_constants(config, store, request, errors)
    if "packets" in request.domains:
        _validate_packets(store, errors)

    _validate_configured_files_and_blocks(config, request, errors)

    if errors:
        raise ValidationError(errors)


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
