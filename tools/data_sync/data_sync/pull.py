"""Reverse sync support."""

from __future__ import annotations

from typing import Callable, Any

from data_sync.block_io import BlockIOError, extract_block
from data_sync.config import DataSyncConfig, DomainLanguageConfig
from data_sync.constants_store import ConstantsStore, ConstantsStoreError
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


def pull_constants(config: DataSyncConfig, store: TomlStore, language: str) -> None:
    parser = CONSTANTS_PARSERS.get(language)
    if parser is None:
        raise PullError(f"unsupported constants pull language: {language}")

    targets = config.targets_for("constants", language)
    if not targets:
        raise PullError(f"missing config for [constants.{language}]")
    constants_store = _load_constants_store(config)
    for target in targets:
        if not target.enabled:
            raise PullError(f"[constants.{target.language}] is disabled in config")
        for section_name in target.owns:
            parsed = _parse_owned_section(target, section_name, parser)
            try:
                source_path = constants_store.source_path(section_name)
            except ConstantsStoreError as exc:
                raise PullError(str(exc)) from exc

            source_store = store if source_path == store.path else _load_source_store(source_path)

            existing = source_store.constants(section_name)
            existing_names = tuple(name for name, _value in existing.values)
            parsed_names = tuple(name for name, _value in parsed)
            if parsed_names != existing_names:
                raise PullError(
                    f"[{section_name}] pull may only update existing values; "
                    f"expected keys {existing_names}, found {parsed_names}"
                )
            source_store.update_constants(section_name, dict(parsed))
            source_store.write(source_path)


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


def _parse_owned_section(
    target: DomainLanguageConfig,
    section_name: str,
    parser: Parser,
) -> tuple[tuple[str, Any], ...]:
    parsed_values: tuple[tuple[str, Any], ...] | None = None
    for path in target.files:
        try:
            text = path.read_text(encoding="utf-8")
        except FileNotFoundError as exc:
            raise PullError(f"configured constants file does not exist: {path}") from exc
        except OSError as exc:
            raise PullError(f"failed to read constants file {path}: {exc}") from exc

        try:
            block = extract_block(text, section_name)
            current_values = parser(block)
        except BlockIOError as exc:
            raise PullError(f"{path}: {exc}") from exc
        except ConstantsParseError as exc:
            raise PullError(f"{path} [{section_name}]: {exc}") from exc

        if parsed_values is not None and current_values != parsed_values:
            raise PullError(f"[{section_name}] has conflicting values across configured files")
        parsed_values = current_values

    if parsed_values is None:
        raise PullError(f"[{section_name}] has no configured source files")
    return parsed_values
