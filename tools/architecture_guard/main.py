"""Rule loading and evaluation for the architecture guard."""

from __future__ import annotations

import argparse
import fnmatch
import os
import re
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import tomlkit


@dataclass(frozen=True)
class Finding:
    rule_id: str
    path: str | None
    line: int | None
    message: str

    def format(self) -> str:
        location = self.path or "<repository>"
        if self.line is not None:
            location += f":{self.line}"
        return f"{self.rule_id}: {location}: {self.message}"


def run_guard(root: Path, config: Path | None = None) -> list[Finding]:
    root = root.resolve()
    config = config or root / "tools" / "architecture_guard" / "rules.toml"
    document = tomlkit.loads(config.read_text(encoding="utf-8"))
    findings: list[Finding] = []
    config_relative = config.resolve().relative_to(root).as_posix()
    for rule in document.get("content_rules", []):
        findings.extend(_check_content(root, rule, {config_relative}))
    for rule in document.get("path_rules", []):
        findings.extend(_check_path(root, rule))
    return findings


def _check_content(root: Path, rule: Any, ignored: set[str]) -> list[Finding]:
    includes = list(rule.get("include", ["**/*"]))
    excludes = list(rule.get("exclude", []))
    pattern = str(rule["pattern"])
    matcher = re.compile(pattern) if rule.get("kind", "literal") == "regex" else None
    message = str(rule.get("message", f"forbidden pattern found: {pattern}"))
    findings: list[Finding] = []
    for path in _files(root):
        relative = path.relative_to(root).as_posix()
        if relative in ignored or not _matches(relative, includes) or _matches(relative, excludes):
            continue
        for number, line in enumerate(path.read_text(encoding="utf-8", errors="replace").splitlines(), 1):
            matched = bool(matcher.search(line)) if matcher else pattern in line
            if matched:
                findings.append(Finding(str(rule["id"]), relative, number, message))
    return findings


def _check_path(root: Path, rule: Any) -> list[Finding]:
    pattern = str(rule["pattern"])
    exists = any(_matches(path.relative_to(root).as_posix(), [pattern]) for path in _files(root))
    exists = exists or any(_matches(path.relative_to(root).as_posix(), [pattern]) for path in _dirs(root))
    kind = str(rule.get("kind", "forbidden"))
    violation = not exists if kind == "required" else exists
    if not violation:
        return []
    default = "required path is missing" if kind == "required" else "forbidden path exists"
    return [Finding(str(rule["id"]), None, None, str(rule.get("message", f"{default}: {pattern}")))]


def _files(root: Path):
    return _walk(root, files=True)


def _dirs(root: Path):
    return _walk(root, files=False)


def _walk(root: Path, files: bool):
    excluded = {".git", ".worktrees", ".workingtrees", "__pycache__", ".godot", "node_modules"}
    for current, dirnames, filenames in os.walk(root):
        dirnames[:] = [name for name in dirnames if name not in excluded]
        names = filenames if files else dirnames
        yield from (Path(current) / name for name in names)


def _matches(path: str, patterns: list[str]) -> bool:
    return any(
        fnmatch.fnmatch(path, pattern)
        or fnmatch.fnmatch(path, pattern.lstrip("./"))
        or (pattern.startswith("**/") and fnmatch.fnmatch(path, pattern[3:]))
        for pattern in patterns
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Check repository architecture rules")
    parser.add_argument("root", nargs="?", type=Path, default=Path.cwd())
    parser.add_argument("--config", type=Path)
    args = parser.parse_args(argv)
    findings = run_guard(args.root, args.config)
    for finding in findings:
        print(finding.format())
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main())
