from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.config import (
    DEFAULT_CONSTANTS_SCAN,
    ConfigError,
    DataSyncConfig,
    DomainLanguageConfig,
    ScanConfig,
    load_config,
)


def write_config(tmp_path: Path, body: str) -> Path:
    path = tmp_path / "config.toml"
    path.write_text(body, encoding="utf-8")
    return path


def valid_config() -> str:
    return """
[sot]
path = "shared/game_data.toml"

[constants.scan]
include = ["client/scripts/**/*.gd", "services/game-server/**/*.go"]
exclude = ["**/.godot/**", "**/vendor/**"]

[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
owns = ["packets"]

[packets.gds]
files = ["client/scripts/packets.gd"]
sections = ["packets"]
owns = []

[packets.ts]
files = ["services/api-server/src/packets.ts"]
sections = ["packets"]
owns = []

[drop_tables.go]
files = ["services/game-server/internal/game/drops/drop_tables.go"]
sections = ["drop_tables.basicasteroids"]
owns = []
outputs = ["server_drop_tables"]
""".strip()


def test_loads_valid_config(tmp_path: Path) -> None:
    config_path = write_config(tmp_path, valid_config())

    config = load_config(config_path)

    assert config.path == config_path.resolve()
    assert config.root == tmp_path.resolve()
    assert config.sot_path("constants") == tmp_path / "shared/game_data.toml"
    assert config.sot_path("packets") == tmp_path / "shared/game_data.toml"
    assert config.target("packets", "go").files == (tmp_path / "services/game-server/internal/network/packets.go",)
    assert config.constants_scan == ScanConfig(
        include=("client/scripts/**/*.gd", "services/game-server/**/*.go"),
        exclude=("**/.godot/**", "**/vendor/**"),
    )


def test_loads_constants_scan_config(tmp_path: Path) -> None:
    config_path = write_config(
        tmp_path,
        valid_config()
        + """

[constants.scan]
include = ["client/scripts/constants/**/*.gd", "services/game-server/internal/**/*.go"]
exclude = ["**/.godot/**", "**/generated/**"]
""",
    )

    config = load_config(config_path)

    assert config.constants_scan == ScanConfig(
        include=("client/scripts/constants/**/*.gd", "services/game-server/internal/**/*.go"),
        exclude=("**/.godot/**", "**/generated/**"),
    )


def test_constants_scan_config_loads_without_per_output_constants_tables(tmp_path: Path) -> None:
    config_path = write_config(
        tmp_path,
        """
[sot.constants]
path = "shared/game_data.toml"

[constants.scan]
include = ["client/scripts/**/*.gd", "services/game-server/**/*.go"]
exclude = ["**/.godot/**", "**/vendor/**"]
""".strip(),
    )

    config = load_config(config_path)

    assert config.constants_scan.include == (
        "client/scripts/**/*.gd",
        "services/game-server/**/*.go",
    )
    assert config.constants_scan.exclude == (
        "**/.godot/**",
        "**/vendor/**",
    )


def test_targets_for_can_return_multiple_targets(tmp_path: Path) -> None:
    first = DomainLanguageConfig(
        domain="constants",
        language="go",
        label="constants.go",
        files=(tmp_path / "services/game-server/internal/constants/constants.go",),
        sections=("constants.server.damage",),
        owns=("constants.server.damage",),
    )
    second = DomainLanguageConfig(
        domain="constants",
        language="go",
        label="weapons.go",
        files=(tmp_path / "services/game-server/internal/constants/weapons.go",),
        sections=("constants.server.weapons.basic_cannon",),
        owns=("constants.server.weapons.basic_cannon",),
    )
    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={"constants": (tmp_path / "shared/constants.toml",)},
        targets_by_domain_language={("constants", "go"): (first, second)},
        constants_scan=DEFAULT_CONSTANTS_SCAN,
    )

    assert config.target("constants", "go") is first
    assert config.targets_for("constants", "go") == (first, second)
    assert config.filter_targets(("constants",), ("go",)) == (first, second)
    assert config.enabled_languages("constants") == ("go",)


def test_enabled_languages_ignores_missing_language_entries(tmp_path: Path) -> None:
    target = DomainLanguageConfig(
        domain="constants",
        language="go",
        label="constants.go",
        files=(tmp_path / "services/game-server/internal/constants/constants.go",),
        sections=("constants.server.damage",),
        owns=("constants.server.damage",),
    )
    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={"constants": (tmp_path / "shared/constants.toml",)},
        targets_by_domain_language={("constants", "go"): (target,)},
        constants_scan=DEFAULT_CONSTANTS_SCAN,
    )

    assert config.enabled_languages("constants") == ("go",)


def test_sot_override(tmp_path: Path) -> None:
    config_path = write_config(tmp_path, valid_config())

    config = load_config(config_path, "custom/source.toml")

    assert config.sot_path("constants") == tmp_path / "custom/source.toml"
    assert config.sot_path("packets") == tmp_path / "custom/source.toml"


def test_loads_per_domain_sot_paths(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        """
[sot]
path = "shared/game_data.toml"
""".strip(),
        """
[sot.constants]
path = "shared/game_data.toml"

[sot.packets]
path = "shared/packets/packets.toml"
""".strip(),
    )
    config_path = write_config(tmp_path, config_text)

    config = load_config(config_path)

    assert config.sot_path("constants") == tmp_path / "shared/game_data.toml"
    assert config.sot_path("packets") == tmp_path / "shared/packets/packets.toml"


def test_loads_per_domain_sot_paths_arrays(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        """
[sot]
path = "shared/game_data.toml"
""".strip(),
        """
[sot.constants]
paths = ["shared/game_data.toml", "shared/game_data.override.toml"]

[sot.packets]
paths = ["shared/packets/packets.toml", "shared/packets/packets.override.toml"]
""".strip(),
    )
    config_path = write_config(tmp_path, config_text)

    config = load_config(config_path)

    assert config.sot_paths("constants") == (
        tmp_path / "shared/game_data.toml",
        tmp_path / "shared/game_data.override.toml",
    )
    assert config.sot_paths("packets") == (
        tmp_path / "shared/packets/packets.toml",
        tmp_path / "shared/packets/packets.override.toml",
    )


def test_invalid_domain_paths_type_raises_clear_error(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        """
[sot]
path = "shared/game_data.toml"
""".strip(),
        """
[sot.constants]
paths = "not-a-list"
""".strip(),
    )
    config_path = write_config(tmp_path, config_text)

    with pytest.raises(ConfigError, match=r"\[sot.constants\]\.paths must be an array"):
        load_config(config_path)


def test_missing_config_raises_clear_error(tmp_path: Path) -> None:
    with pytest.raises(ConfigError, match="config file does not exist"):
        load_config(tmp_path / "missing.toml")


def test_malformed_config_raises_clear_error(tmp_path: Path) -> None:
    config_path = write_config(tmp_path, "[sot\npath = ")

    with pytest.raises(ConfigError, match="failed to parse TOML config|expected TOML key/value"):
        load_config(config_path)


def test_missing_required_packet_key_raises_clear_error(tmp_path: Path) -> None:
    config_path = write_config(
        tmp_path,
        """
[sot]
path = "shared/game_data.toml"

[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
""".strip(),
    )

    with pytest.raises(ConfigError, match=r"\[packets\.go\] missing required key\(s\): owns"):
        load_config(config_path)


def test_missing_required_packet_table_key_raises_clear_error(tmp_path: Path) -> None:
    config_text = valid_config().replace('owns = ["packets"]', "", 1)
    config_path = write_config(tmp_path, config_text)

    with pytest.raises(ConfigError, match=r"\[packets\.go\] missing required key"):
        load_config(config_path)


def test_incomplete_constants_tables_are_ignored(tmp_path: Path) -> None:
    config_path = write_config(
        tmp_path,
        """
[sot.constants]
path = "shared/game_data.toml"

[constants.go]
files = ["services/game-server/internal/game/constants.go"]
sections = ["constants.gameplay"]

[constants.scan]
include = ["client/scripts/**/*.gd", "services/game-server/**/*.go"]
exclude = []

[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
owns = ["packets"]
""".strip(),
    )

    config = load_config(config_path)

    assert config.constants_scan.include == (
        "client/scripts/**/*.gd",
        "services/game-server/**/*.go",
    )
    with pytest.raises(ConfigError, match=r"missing config for \[constants\.go\]"):
        config.targets_for("constants", "go")


def test_filter_targets_uses_requested_domains_and_languages(tmp_path: Path) -> None:
    config_path = write_config(tmp_path, valid_config())
    config = load_config(config_path)

    targets = config.filter_targets(("packets",), ("ts",))

    assert [(target.domain, target.language) for target in targets] == [("packets", "ts")]


def test_packet_target_supports_optional_outputs_ids_in_order(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        """
[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
owns = ["packets"]
""".strip(),
        """
[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
owns = ["packets"]
outputs = ["server_entities_packets", "server_game_packets"]
""".strip(),
    )
    config_path = write_config(tmp_path, config_text)
    config = load_config(config_path)

    packets_go = config.target("packets", "go")
    assert packets_go.files == (tmp_path / "services/game-server/internal/network/packets.go",)
    assert packets_go.sections == ("packets",)
    assert packets_go.owns == ("packets",)
    assert packets_go.outputs == ("server_entities_packets", "server_game_packets")

    assert config.target("packets", "go").files == (tmp_path / "services/game-server/internal/network/packets.go",)
    assert config.target("drop_tables", "go").files == (tmp_path / "services/game-server/internal/game/drops/drop_tables.go",)


def test_packet_target_outputs_must_be_list_of_non_empty_strings(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        """
[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
owns = ["packets"]
""".strip(),
        """
[packets.go]
files = ["services/game-server/internal/network/packets.go"]
sections = ["packets"]
owns = ["packets"]
outputs = ["server_entities_packets", ""]
""".strip(),
    )
    config_path = write_config(tmp_path, config_text)

    with pytest.raises(ConfigError, match=r"\[packets.go\]\.outputs must contain only non-empty strings"):
        load_config(config_path)
