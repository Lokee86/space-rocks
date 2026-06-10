"""Reverse sync support."""

from __future__ import annotations

from typing import Callable, Any

from data_sync.block_io import BlockIOError, extract_block
from data_sync.config import ConfigError, DataSyncConfig
from data_sync.constants_store import ConstantsStore, ConstantsStoreError
from data_sync.discovery import discover_constants_files
from data_sync.parsers import gds_constants, go_constants, ts_constants
from data_sync.parsers.go_constants import ConstantsParseError
from data_sync.toml_store import TomlStore, TomlStoreError


class PullError(Exception):
    """Raised when reverse sync cannot complete safely."""


Parser = Callable[[str], tuple[tuple[str, Any], ...]]


CONSTANTS_PARSERS: dict[str, Parser] = {
    "go": go_constants.parse_constants,
    "gds": gds_constants.parse_constants,
    "ts": ts_constants.parse_constants,
}


def pull_constants(config: DataSyncConfig, language: str) -> None:
    parser = CONSTANTS_PARSERS.get(language)
    if parser is None:
        raise PullError(f"unsupported constants pull language: {language}")

    constants_store = _load_constants_store(config)
    targets = _configured_constants_targets(config, language)
    if targets:
        parsed_by_section = _parse_configured_sections(targets, parser)
    else:
        discovered_files = discover_constants_files(config, (language,))
        if not discovered_files:
            return
        parsed_by_section = _parse_discovered_sections(discovered_files, parser)
    source_stores: dict[object, TomlStore] = {}
    touched_paths: list[object] = []

    for section_name, parsed in parsed_by_section.items():
        try:
            source_path = constants_store.source_path(section_name)
        except ConstantsStoreError as exc:
            raise PullError(str(exc)) from exc

        source_store = source_stores.get(source_path)
        if source_store is None:
            source_store = _load_source_store(source_path)
            source_stores[source_path] = source_store
            touched_paths.append(source_path)

        existing = source_store.constants(section_name)
        existing_names = tuple(name for name, _value in existing.values)
        parsed_names = tuple(name for name, _value in parsed)
        if parsed_names != existing_names:
            raise PullError(
                f"[{section_name}] pull may only update existing values; "
                f"expected keys {existing_names}, found {parsed_names}"
            )
        source_store.update_constants(section_name, dict(parsed))

    for source_path in touched_paths:
        source_stores[source_path].write(source_path)


def _configured_constants_targets(config: DataSyncConfig, language: str):
    try:
        return config.targets_for("constants", language)
    except ConfigError:
        return ()


def _load_constants_store(config: DataSyncConfig) -> ConstantsStore:
    sot_paths = config.sot_paths("constants")
    if not sot_paths:
        raise PullError("missing constants SoT paths")
    try:
        return ConstantsStore.load(sot_paths)
    except ConstantsStoreError as exc:
        raise PullError(str(exc)) from exc


def _load_source_store(path) -> TomlStore:
    try:
        return TomlStore.load(path)
    except TomlStoreError as exc:
        raise PullError(str(exc)) from exc


def _parse_configured_sections(
    targets,
    parser: Parser,
) -> dict[str, tuple[tuple[str, Any], ...]]:
    parsed_by_section: dict[str, tuple[tuple[str, Any], ...]] = {}
    section_owners: set[str] = set()

    for target in targets:
        texts: dict[object, str] = {}
        for path in target.files:
            try:
                texts[path] = path.read_text(encoding="utf-8")
            except FileNotFoundError as exc:
                raise PullError(f"configured constants file does not exist: {path}") from exc
            except OSError as exc:
                raise PullError(f"failed to read constants file {path}: {exc}") from exc

        for section_name in target.sections:
            if section_name in section_owners:
                raise PullError(f"[{section_name}] is owned by multiple targets")

            current_values = None
            current_source = None
            for path, text in texts.items():
                try:
                    block = extract_block(text, section_name)
                except BlockIOError as exc:
                    message = str(exc)
                    if "missing data-sync block" in message:
                        continue
                    raise PullError(f"{path}: {exc}") from exc

                if current_source is not None and path != current_source:
                    raise PullError(f"[{section_name}] is owned by multiple targets")

                try:
                    current_values = parser(block)
                except ConstantsParseError as exc:
                    raise PullError(f"{path} [{section_name}]: {exc}") from exc
                current_source = path

            if current_values is None:
                raise PullError(f"configured constants target missing section [{section_name}]")

            section_owners.add(section_name)
            previous_values = parsed_by_section.get(section_name)
            if previous_values is not None and current_values != previous_values:
                raise PullError(f"[{section_name}] has conflicting values across configured targets")
            parsed_by_section[section_name] = current_values

    return parsed_by_section


def _parse_discovered_sections(
    discovered_files,
    parser: Parser,
) -> dict[str, tuple[tuple[str, Any], ...]]:
    parsed_by_section: dict[str, tuple[tuple[str, Any], ...]] = {}
    for discovered in discovered_files:
        path = discovered.path
        try:
            text = path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise PullError(f"discovered constants file does not exist: {path}") from exc
        except OSError as exc:
            raise PullError(f"failed to read constants file {path}: {exc}") from exc

        for section_name in discovered.sections:
            try:
                block = extract_block(text, section_name)
                current_values = parser(block)
            except BlockIOError as exc:
                raise PullError(f"{path}: {exc}") from exc
            except ConstantsParseError as exc:
                raise PullError(f"{path} [{section_name}]: {exc}") from exc

            previous_values = parsed_by_section.get(section_name)
            if previous_values is not None and current_values != previous_values:
                raise PullError(f"[{section_name}] has conflicting values across discovered files")
            parsed_by_section[section_name] = current_values
    return parsed_by_section
