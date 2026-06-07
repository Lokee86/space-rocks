"""Read-only constants wrapper across one or more TOML sources."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

from data_sync.model.constants import ConstantSection
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
        source = self._source_for_section(section_name)
        return source.store.constants(section_name)

    def source_path(self, section_name: str) -> Path:
        return self._source_for_section(section_name).path

    def _source_for_section(self, section_name: str) -> _StoreSource:
        matches = []
        for source in self._sources:
            if source.store.has_section(section_name):
                matches.append(source)

        if not matches:
            raise ConstantsStoreError(f"missing TOML section [{section_name}]")
        if len(matches) > 1:
            raise ConstantsStoreError(
                f"duplicate constants source for [{section_name}]: "
                + ", ".join(str(source.path) for source in matches)
            )
        return matches[0]
