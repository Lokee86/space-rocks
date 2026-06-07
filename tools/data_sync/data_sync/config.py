"""Configuration loading for data sync."""

from __future__ import annotations

import ast
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Mapping

from data_sync.cli import DOMAINS, LANGUAGES


DEFAULT_CONFIG_PATH = Path(__file__).resolve().parents[1] / "config.toml"
DEFAULT_SOT_PATHS = {
    "constants": (
        "shared/constants/server_constants.toml",
        "shared/constants/server_entities.toml",
        "shared/constants/client/presentation.toml",
        "shared/constants/client/shell.toml",
        "shared/constants/client/lobby.toml",
    ),
    "packets": (
        "shared/packets/outputs.toml",
        "shared/packets/gameplay.toml",
        "shared/packets/debug.toml",
        "shared/packets/lobby.toml",
    ),
    "drop_tables": (
        "shared/drop_tables/basicasteroids.toml",
    ),
}
REQUIRED_DOMAIN_KEYS = ("files", "sections", "owns")


class ConfigError(Exception):
    """Raised when a data sync config cannot be loaded or validated."""


@dataclass(frozen=True)
class DomainLanguageConfig:
    domain: str
    language: str
    label: str
    files: tuple[Path, ...]
    sections: tuple[str, ...]
    owns: tuple[str, ...]
    outputs: tuple[str, ...] = ()
    enabled: bool = True

    def receives_section(self, section: str) -> bool:
        return section in self.sections

    def owns_section(self, section: str) -> bool:
        return section in self.owns

    def display_name(self) -> str:
        return self.label or f"{self.domain}.{self.language}"


@dataclass(frozen=True)
class DataSyncConfig:
    path: Path
    root: Path
    sot_paths_by_domain: Mapping[str, tuple[Path, ...]]
    targets_by_domain_language: Mapping[tuple[str, str], tuple[DomainLanguageConfig, ...]]

    def sot_paths(self, domain: str) -> tuple[Path, ...]:
        try:
            return self.sot_paths_by_domain[domain]
        except KeyError as exc:
            raise ConfigError(f"missing SoT path for domain: {domain}") from exc

    def sot_path(self, domain: str) -> Path:
        paths = self.sot_paths(domain)
        if len(paths) != 1:
            raise ConfigError(
                f"domain {domain!r} has multiple SoT paths; use sot_paths({domain!r}) instead"
            )
        return paths[0]

    def target(self, domain: str, language: str) -> DomainLanguageConfig:
        targets = self.targets_for(domain, language)
        if not targets:
            raise ConfigError(f"missing config for [{domain}.{language}]")
        return targets[0]

    def targets_for(self, domain: str, language: str) -> tuple[DomainLanguageConfig, ...]:
        try:
            return self.targets_by_domain_language[(domain, language)]
        except KeyError as exc:
            raise ConfigError(f"missing config for [{domain}.{language}]") from exc

    def enabled_languages(self, domain: str) -> tuple[str, ...]:
        if domain == "drop_tables" and not any(
            key_domain == domain and key_language == "go"
            for key_domain, key_language in self.targets_by_domain_language
        ):
            return ()
        return tuple(
            language
            for language in LANGUAGES
            if any(
                target.enabled
                for target in self.targets_by_domain_language.get((domain, language), ())
            )
        )

    def filter_targets(
        self,
        domains: tuple[str, ...] | list[str],
        languages: tuple[str, ...] | list[str],
    ) -> tuple[DomainLanguageConfig, ...]:
        return tuple(
            target
            for domain in domains
            for language in languages
            for target in self.targets_for(domain, language)
        )


def load_config(config_path: Path | str | None = None, sot_override: Path | str | None = None) -> DataSyncConfig:
    resolved_config_path = _resolve_config_path(config_path)
    raw = _load_toml_file(resolved_config_path)
    root = _resolve_config_root(resolved_config_path)

    sot_values = _read_sot_paths(raw)
    if sot_override is not None:
        sot_values = {domain: (str(sot_override),) for domain in DOMAINS}

    targets: dict[tuple[str, str], list[DomainLanguageConfig]] = {}
    for domain in DOMAINS:
        if domain == "drop_tables" and domain not in raw:
            continue
        domain_table = raw.get(domain)
        if domain_table is None:
            continue
        if not isinstance(domain_table, Mapping):
            raise ConfigError(f"missing required config table [{domain}]")
        domain_languages = ("go",) if domain == "drop_tables" else LANGUAGES
        for language in domain_languages:
            table = domain_table.get(language)
            if table is None:
                continue
            if not isinstance(table, Mapping):
                raise ConfigError(f"missing required config table [{domain}.{language}]")
            targets.setdefault((domain, language), []).append(
                _load_domain_language_config(root, domain, language, table)
            )
    for target in _discover_constants_outputs(root, raw):
        targets.setdefault((target.domain, target.language), []).append(target)

    _validate_constants_ownership(
        target
        for target_list in targets.values()
        for target in target_list
    )

    return DataSyncConfig(
        path=resolved_config_path,
        root=root,
        sot_paths_by_domain={
            domain: tuple(_resolve_path(root, value) for value in values)
            for domain, values in sot_values.items()
        },
        targets_by_domain_language={
            key: tuple(value)
            for key, value in targets.items()
        },
    )


def _resolve_config_path(config_path: Path | str | None) -> Path:
    path = Path(config_path) if config_path is not None else DEFAULT_CONFIG_PATH
    path = path.expanduser()
    if not path.is_absolute():
        path = (Path.cwd() / path).resolve()
    else:
        path = path.resolve()
    if not path.exists():
        raise ConfigError(f"config file does not exist: {path}")
    if not path.is_file():
        raise ConfigError(f"config path is not a file: {path}")
    return path


def _load_toml_file(path: Path) -> Mapping[str, Any]:
    try:
        with path.open("rb") as handle:
            return _toml_load(handle)
    except ConfigError:
        raise
    except Exception as exc:
        raise ConfigError(f"failed to parse TOML config {path}: {exc}") from exc


def _toml_load(handle: Any) -> Mapping[str, Any]:
    try:
        import tomllib

        return tomllib.load(handle)
    except ModuleNotFoundError:
        pass

    try:
        import tomli

        return tomli.load(handle)
    except ModuleNotFoundError:
        pass

    text = handle.read().decode("utf-8")
    return _parse_minimal_toml(text)


def _parse_minimal_toml(text: str) -> dict[str, Any]:
    result: dict[str, Any] = {}
    current: dict[str, Any] | None = None

    for line_number, raw_line in enumerate(text.splitlines(), start=1):
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("[") and line.endswith("]"):
            section_name = line[1:-1].strip()
            if not section_name:
                raise ConfigError(f"empty table name on line {line_number}")
            current = result
            for part in section_name.split("."):
                if not part:
                    raise ConfigError(f"invalid table name on line {line_number}")
                next_table = current.setdefault(part, {})
                if not isinstance(next_table, dict):
                    raise ConfigError(f"table conflicts with value on line {line_number}")
                current = next_table
            continue
        if current is None or "=" not in line:
            raise ConfigError(f"expected TOML key/value on line {line_number}")
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip()
        if not key:
            raise ConfigError(f"empty key on line {line_number}")
        try:
            parsed_value = ast.literal_eval(value)
        except (SyntaxError, ValueError) as exc:
            raise ConfigError(f"unsupported TOML value on line {line_number}: {value}") from exc
        current[key] = parsed_value

    return result


def _resolve_config_root(config_path: Path) -> Path:
    cwd_repo_root = _find_repo_root(Path.cwd())
    if cwd_repo_root is not None:
        try:
            config_path.relative_to(cwd_repo_root)
            return cwd_repo_root
        except ValueError:
            pass
    return config_path.parent


def _find_repo_root(start: Path) -> Path | None:
    for candidate in (start, *start.parents):
        if (candidate / ".git").exists():
            return candidate
    return None


def _read_sot_paths(raw: Mapping[str, Any]) -> dict[str, tuple[str, ...]]:
    sot_table = raw.get("sot", {})
    if sot_table is None:
        return dict(DEFAULT_SOT_PATHS)
    if not isinstance(sot_table, Mapping):
        raise ConfigError("[sot] must be a table")

    legacy_value = sot_table.get("path")
    if legacy_value is not None:
        if not isinstance(legacy_value, str) or not legacy_value:
            raise ConfigError("[sot].path must be a non-empty string")
        return {domain: (legacy_value,) for domain in DOMAINS}

    paths: dict[str, tuple[str, ...]] = {}
    for domain in DOMAINS:
        if domain == "drop_tables" and domain not in sot_table:
            continue
        domain_table = sot_table.get(domain, {})
        if domain_table is None:
            domain_table = {}
        if not isinstance(domain_table, Mapping):
            raise ConfigError(f"[sot.{domain}] must be a table")

        legacy_domain_path = domain_table.get("path")
        domain_paths = domain_table.get("paths")
        if legacy_domain_path is not None and domain_paths is not None:
            raise ConfigError(f"[sot.{domain}] must not specify both path and paths")

        if domain_paths is not None:
            if not isinstance(domain_paths, list):
                raise ConfigError(f"[sot.{domain}].paths must be an array of non-empty strings")
            if not domain_paths or not all(isinstance(item, str) and item for item in domain_paths):
                raise ConfigError(f"[sot.{domain}].paths must be an array of non-empty strings")
            paths[domain] = tuple(domain_paths)
            continue

        if legacy_domain_path is not None:
            if not isinstance(legacy_domain_path, str) or not legacy_domain_path:
                raise ConfigError(f"[sot.{domain}].path must be a non-empty string")
            paths[domain] = (legacy_domain_path,)
            continue

        paths[domain] = DEFAULT_SOT_PATHS[domain]
    return paths


def _require_table(raw: Mapping[str, Any], key: str, label: str | None = None) -> Mapping[str, Any]:
    table = raw.get(key)
    label = label or key
    if not isinstance(table, Mapping):
        raise ConfigError(f"missing required config table [{label}]")
    return table


def _load_domain_language_config(
    root: Path,
    domain: str,
    language: str,
    table: Mapping[str, Any],
) -> DomainLanguageConfig:
    label = f"{domain}.{language}"
    missing = [key for key in REQUIRED_DOMAIN_KEYS if key not in table]
    if missing:
        raise ConfigError(f"[{label}] missing required key(s): {', '.join(missing)}")

    enabled = _read_bool(table.get("enabled", True), f"[{label}].enabled")
    files = _read_string_list(table["files"], f"[{label}].files")
    sections = _read_string_list(table["sections"], f"[{label}].sections")
    owns = _read_string_list(table["owns"], f"[{label}].owns")
    outputs: tuple[str, ...] = ()
    if domain in {"packets", "drop_tables"}:
        outputs = tuple(_read_string_list(table.get("outputs", []), f"[{label}].outputs"))

    if enabled and domain != "drop_tables" and not files:
        raise ConfigError(f"[{label}].files must not be empty")
    if enabled and domain != "drop_tables" and not sections:
        raise ConfigError(f"[{label}].sections must not be empty")

    unknown_owns = [section for section in owns if section not in sections]
    if unknown_owns:
        raise ConfigError(
            f"[{label}].owns contains section(s) not listed in sections: {', '.join(unknown_owns)}"
        )

    return DomainLanguageConfig(
        domain=domain,
        language=language,
        label=label,
        files=tuple(_resolve_path(root, value) for value in files),
        sections=tuple(sections),
        owns=tuple(owns),
        outputs=outputs,
        enabled=enabled,
    )


def _read_string_list(value: Any, label: str) -> list[str]:
    if not isinstance(value, list):
        raise ConfigError(f"{label} must be a list of strings")
    if not all(isinstance(item, str) and item for item in value):
        raise ConfigError(f"{label} must contain only non-empty strings")
    return value


def _read_bool(value: Any, label: str) -> bool:
    if not isinstance(value, bool):
        raise ConfigError(f"{label} must be a boolean")
    return value


def _resolve_path(root: Path, value: str | Path) -> Path:
    path = Path(value).expanduser()
    if path.is_absolute():
        return path.resolve()
    return (root / path).resolve()


def _validate_constants_ownership(targets: Any) -> None:
    owners: dict[str, str] = {}
    for target in targets:
        domain = target.domain
        if domain != "constants":
            continue
        for section in target.owns:
            previous = owners.get(section)
            if previous is not None:
                raise ConfigError(
                    f"constants section {section!r} is owned by multiple targets: {previous}, {target.display_name()}"
                )
            owners[section] = target.display_name()


def _discover_constants_outputs(root: Path, raw: Mapping[str, Any]) -> tuple[DomainLanguageConfig, ...]:
    discovered: list[DomainLanguageConfig] = []
    for name, table in raw.items():
        if name in DOMAINS or name == "sot":
            continue
        if not isinstance(table, Mapping):
            continue
        for language in LANGUAGES:
            if language not in table:
                continue
            language_table = table.get(language)
            if not isinstance(language_table, Mapping):
                continue
            if not _looks_like_constants_output(language_table):
                continue
            discovered.append(_load_domain_language_config(root, "constants", language, language_table))
    return tuple(discovered)


def _looks_like_constants_output(table: Mapping[str, Any]) -> bool:
    for key in REQUIRED_DOMAIN_KEYS:
        if key not in table:
            return False
    sections = table.get("sections")
    if not isinstance(sections, list):
        return False
    return all(isinstance(section, str) and section.startswith("constants.") for section in sections)
