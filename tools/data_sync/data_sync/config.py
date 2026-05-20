"""Configuration loading for data sync."""

from __future__ import annotations

import ast
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Mapping

from data_sync.cli import DOMAINS, LANGUAGES


DEFAULT_CONFIG_PATH = Path(__file__).resolve().parents[1] / "config.toml"
DEFAULT_SOT_PATH = "shared/game_data.toml"
REQUIRED_DOMAIN_KEYS = ("files", "sections", "owns")


class ConfigError(Exception):
    """Raised when a data sync config cannot be loaded or validated."""


@dataclass(frozen=True)
class DomainLanguageConfig:
    domain: str
    language: str
    files: tuple[Path, ...]
    sections: tuple[str, ...]
    owns: tuple[str, ...]
    enabled: bool = True

    def receives_section(self, section: str) -> bool:
        return section in self.sections

    def owns_section(self, section: str) -> bool:
        return section in self.owns


@dataclass(frozen=True)
class DataSyncConfig:
    path: Path
    root: Path
    sot_path: Path
    targets: Mapping[tuple[str, str], DomainLanguageConfig]

    def target(self, domain: str, language: str) -> DomainLanguageConfig:
        try:
            return self.targets[(domain, language)]
        except KeyError as exc:
            raise ConfigError(f"missing config for [{domain}.{language}]") from exc

    def enabled_languages(self, domain: str) -> tuple[str, ...]:
        return tuple(
            language
            for language in LANGUAGES
            if self.target(domain, language).enabled
        )

    def filter_targets(
        self,
        domains: tuple[str, ...] | list[str],
        languages: tuple[str, ...] | list[str],
    ) -> tuple[DomainLanguageConfig, ...]:
        return tuple(self.target(domain, language) for domain in domains for language in languages)


def load_config(config_path: Path | str | None = None, sot_override: Path | str | None = None) -> DataSyncConfig:
    resolved_config_path = _resolve_config_path(config_path)
    raw = _load_toml_file(resolved_config_path)
    root = _resolve_config_root(resolved_config_path)

    sot_value = _read_sot_path(raw)
    if sot_override is not None:
        sot_value = str(sot_override)

    targets: dict[tuple[str, str], DomainLanguageConfig] = {}
    for domain in DOMAINS:
        domain_table = _require_table(raw, domain)
        for language in LANGUAGES:
            table = _require_table(domain_table, language, f"{domain}.{language}")
            targets[(domain, language)] = _load_domain_language_config(root, domain, language, table)

    _validate_constants_ownership(targets)

    return DataSyncConfig(
        path=resolved_config_path,
        root=root,
        sot_path=_resolve_path(root, sot_value),
        targets=targets,
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


def _read_sot_path(raw: Mapping[str, Any]) -> str:
    sot_table = raw.get("sot", {})
    if sot_table is None:
        return DEFAULT_SOT_PATH
    if not isinstance(sot_table, Mapping):
        raise ConfigError("[sot] must be a table")
    value = sot_table.get("path", DEFAULT_SOT_PATH)
    if not isinstance(value, str) or not value:
        raise ConfigError("[sot].path must be a non-empty string")
    return value


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

    if enabled and not files:
        raise ConfigError(f"[{label}].files must not be empty")
    if enabled and not sections:
        raise ConfigError(f"[{label}].sections must not be empty")

    unknown_owns = [section for section in owns if section not in sections]
    if unknown_owns:
        raise ConfigError(
            f"[{label}].owns contains section(s) not listed in sections: {', '.join(unknown_owns)}"
        )

    return DomainLanguageConfig(
        domain=domain,
        language=language,
        files=tuple(_resolve_path(root, value) for value in files),
        sections=tuple(sections),
        owns=tuple(owns),
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


def _validate_constants_ownership(targets: Mapping[tuple[str, str], DomainLanguageConfig]) -> None:
    owners: dict[str, str] = {}
    for (domain, language), target in targets.items():
        if domain != "constants":
            continue
        for section in target.owns:
            previous = owners.get(section)
            if previous is not None:
                raise ConfigError(
                    f"constants section {section!r} is owned by multiple languages: {previous}, {language}"
                )
            owners[section] = language
