"""Drop table model definitions."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class DropTableEntry:
    pickup_type: str
    chance: float
    min_source_size: int
    max_source_size: int


@dataclass(frozen=True)
class DropTable:
    id: str
    source_type: str
    drop_mode: str
    max_drops_per_source: int
    max_active_pickups: int
    entries: tuple[DropTableEntry, ...]


@dataclass(frozen=True)
class DropTablesModel:
    tables: tuple[DropTable, ...]
