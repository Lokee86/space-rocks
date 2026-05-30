"""Read-only constants wrapper across one or more TOML sources."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

from data_sync.model.constants import ConstantSection, ConstantValue
from data_sync.toml_store import TomlStore, TomlStoreError


class ConstantsStoreError(Exception):
    """Raised when constants cannot be loaded across TOML sources."""


@dataclass(frozen=True)
class _StoreSource:
    path: Path
    store: TomlStore


class ConstantsStore:
    def __init__(self, sources: tuple[_StoreSource, ...]) -> None:
        self._sources = sources

    @classmethod
    def load(cls, paths: Iterable[Path | str]) -> "ConstantsStore":
        resolved_paths = tuple(Path(path) for path in paths)
        if not resolved_paths:
            raise ConstantsStoreError("at least one SoT TOML path is required")

        return cls(
            tuple(
                _StoreSource(path=path, store=TomlStore.load(path))
                for path in resolved_paths
            )
        )

    def constants(self, section_name: str) -> ConstantSection:
        if len(self._sources) == 1:
            return self._sources[0].store.constants(section_name)

        merged_values: list[tuple[str, ConstantValue]] = []
        key_sources: dict[str, Path] = {}
        section_found = False

        for source in self._sources:
            try:
                section = source.store.constants(section_name)
            except TomlStoreError as exc:
                if "missing TOML section" in str(exc):
                    continue
                raise

            section_found = True
            for key, value in section.values:
                existing_source = key_sources.get(key)
                if existing_source is not None:
                    raise ConstantsStoreError(
                        f"duplicate constants key in [{section_name}].{key}: "
                        f"{existing_source} and {source.path}"
                    )
                key_sources[key] = source.path
                merged_values.append((key, value))

        if not section_found:
            raise ConstantsStoreError(f"missing TOML section [{section_name}]")

        return ConstantSection(name=section_name, values=tuple(merged_values))
