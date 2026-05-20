"""Command-line parsing for the data sync tool."""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from pathlib import Path
from typing import Sequence


OPERATIONS = ("push", "pull", "diff", "check", "validate")
DOMAINS = ("constants", "packets")
LANGUAGES = ("go", "gds", "ts")


@dataclass(frozen=True)
class CliArgs:
    operation: str
    domains: tuple[str, ...]
    languages: tuple[str, ...]
    config: Path | None
    sot: Path | None


class DataSyncArgumentParser(argparse.ArgumentParser):
    def error(self, message: str) -> None:
        self.print_usage()
        self.exit(2, f"{self.prog}: error: {message}\n")


def build_parser() -> argparse.ArgumentParser:
    parser = DataSyncArgumentParser(
        prog="data-sync",
        description="Sync TOML constants and packet definitions with generated language blocks.",
    )

    for operation in OPERATIONS:
        parser.add_argument(f"-{operation}", action="store_true", help=f"run {operation}")

    for domain in DOMAINS:
        parser.add_argument(f"-{domain}", action="store_true", help=f"include {domain}")

    for language in LANGUAGES:
        parser.add_argument(f"-{language}", action="store_true", help=f"include {language}")

    parser.add_argument("-config", type=Path, help="path to data-sync config TOML")
    parser.add_argument("-sot", type=Path, help="override source-of-truth TOML path")
    return parser


def parse_args(argv: Sequence[str] | None = None) -> CliArgs:
    parser = build_parser()
    namespace = parser.parse_args(argv)

    selected_operations = [name for name in OPERATIONS if getattr(namespace, name)]
    if len(selected_operations) != 1:
        parser.error("select exactly one operation: -push, -pull, -diff, -check, or -validate")

    operation = selected_operations[0]
    domains = tuple(name for name in DOMAINS if getattr(namespace, name))
    languages = tuple(name for name in LANGUAGES if getattr(namespace, name))

    if operation in {"push", "pull", "diff", "check"}:
        if not domains:
            parser.error(f"-{operation} requires at least one domain: -constants and/or -packets")
        if not languages:
            parser.error(f"-{operation} requires at least one language: -go, -gds, and/or -ts")

    if operation == "pull" and len(languages) > 1:
        parser.error("-pull may only use one language at a time")

    return CliArgs(
        operation=operation,
        domains=domains,
        languages=languages,
        config=namespace.config,
        sot=namespace.sot,
    )
