#!/usr/bin/env python3
"""Entrypoint for the data sync CLI."""

from __future__ import annotations

import sys

from data_sync.cli import parse_args
from data_sync.config import ConfigError, DataSyncConfig, load_config
from data_sync.constants_store import ConstantsStore, ConstantsStoreError
from data_sync.constants_sync import ConstantsSyncError, apply_updates, plan_constants_updates, unified_diff
from data_sync.drop_tables_sync import DropTablesSyncError, plan_drop_tables_updates
from data_sync.packets_sync import PacketsSyncError, plan_packets_updates
from data_sync.packet_toml import PacketTomlError, load_packet_schema_files
from data_sync.pull import PullError, pull_constants
from data_sync.toml_store import TomlStore, TomlStoreError
from data_sync.validate import ValidationError, validate


def run(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    try:
        config = load_config(args.config, args.sot)
    except ConfigError as exc:
        print(f"config error: {exc}", file=sys.stderr)
        return 2

    if args.operation == "validate":
        try:
            validate(config, args.domains, args.languages)
        except ValidationError as exc:
            print("validation failed:", file=sys.stderr)
            for error in exc.errors:
                print(f"- {error}", file=sys.stderr)
            return 1
        print("validation passed")
        return 0

    if args.operation == "pull":
        if "packets" in args.domains:
            print(
                "pull error: Packet pull is not supported. Edit packet schema files under shared/packets/.",
                file=sys.stderr,
            )
            return 2
        try:
            pull_constants(config, args.languages[0])
        except (PullError, TomlStoreError) as exc:
            print(f"pull error: {exc}", file=sys.stderr)
            return 1
        return 0

    try:
        updates = []
        if "constants" in args.domains:
            constants_store = ConstantsStore.load(config.sot_paths("constants"))
            updates.extend(plan_constants_updates(config, constants_store, args.languages))
        if "drop_tables" in args.domains:
            updates.extend(plan_drop_tables_updates(config))
        if "packets" in args.domains:
            _ensure_enabled_packet_targets(config, args.languages)
            packet_schema = load_packet_schema_files(config.sot_paths("packets"))
            updates.extend(plan_packets_updates(config, packet_schema, args.languages))
    except (
        ConstantsSyncError,
        ConstantsStoreError,
        DropTablesSyncError,
        PacketsSyncError,
        PacketTomlError,
        TomlStoreError,
    ) as exc:
        print(f"{args.operation} error: {exc}", file=sys.stderr)
        return 1

    if args.operation == "push":
        try:
            apply_updates(updates)
        except OSError as exc:
            print(f"push error: {exc}", file=sys.stderr)
            return 1
        return 0

    if args.operation == "diff":
        diff_text = unified_diff(updates)
        if diff_text:
            print(diff_text, end="")
        return 0

    if args.operation == "check":
        return 0 if all(not update.changed for update in updates) else 1

    print(f"{args.operation}: not implemented yet", file=sys.stderr)
    return 2


def _ensure_enabled_packet_targets(config: DataSyncConfig, languages: tuple[str, ...]) -> None:
    for language in languages:
        target = config.target("packets", language)
        if not target.enabled:
            raise PacketsSyncError(f"[packets.{language}] is disabled in config")


def main() -> None:
    raise SystemExit(run())


if __name__ == "__main__":
    main()
