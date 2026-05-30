from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.config import ConfigError, load_config


def write_config(tmp_path: Path, body: str) -> Path:
    path = tmp_path / "config.toml"
    path.write_text(body, encoding="utf-8")
    return path


def valid_config() -> str:
    return """
[sot]
path = "shared/game_data.toml"

[constants.go]
files = ["services/game-server/internal/game/constants.go"]
sections = ["constants.gameplay", "constants.network"]
owns = ["constants.gameplay", "constants.network"]

[constants.gds]
files = ["client/scripts/constants.gd"]
sections = ["constants.gameplay", "constants.client"]
owns = ["constants.client"]

[constants.ts]
files = ["services/api-server/src/constants.ts"]
sections = ["constants.network", "constants.api"]
owns = ["constants.api"]

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
""".strip()


def test_loads_valid_config(tmp_path: Path) -> None:
    config_path = write_config(tmp_path, valid_config())

    config = load_config(config_path)

    assert config.path == config_path.resolve()
    assert config.root == tmp_path.resolve()
    assert config.sot_path("constants") == tmp_path / "shared/game_data.toml"
    assert config.sot_path("packets") == tmp_path / "shared/game_data.toml"

    go_constants = config.target("constants", "go")
    assert go_constants.sections == ("constants.gameplay", "constants.network")
    assert go_constants.owns == ("constants.gameplay", "constants.network")
    assert go_constants.files == (tmp_path / "services/game-server/internal/game/constants.go",)


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


def test_missing_required_table_raises_clear_error(tmp_path: Path) -> None:
    config_path = write_config(
        tmp_path,
        """
[sot]
path = "shared/game_data.toml"
""".strip(),
    )

    with pytest.raises(ConfigError, match=r"missing required config table \[constants\]"):
        load_config(config_path)


def test_missing_required_key_raises_clear_error(tmp_path: Path) -> None:
    config_text = valid_config().replace('owns = ["constants.api"]', "")
    config_path = write_config(tmp_path, config_text)

    with pytest.raises(ConfigError, match=r"\[constants.ts\] missing required key"):
        load_config(config_path)


def test_constants_ownership_overlap_is_invalid(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        'owns = ["constants.client"]',
        'owns = ["constants.gameplay"]',
    )
    config_path = write_config(tmp_path, config_text)

    with pytest.raises(ConfigError, match="owned by multiple languages"):
        load_config(config_path)


def test_owns_must_be_subset_of_sections(tmp_path: Path) -> None:
    config_text = valid_config().replace(
        'owns = ["constants.api"]',
        'owns = ["constants.missing"]',
    )
    config_path = write_config(tmp_path, config_text)

    with pytest.raises(ConfigError, match="owns contains section"):
        load_config(config_path)


def test_filter_targets_uses_requested_domains_and_languages(tmp_path: Path) -> None:
    config_path = write_config(tmp_path, valid_config())
    config = load_config(config_path)

    targets = config.filter_targets(("packets", "constants"), ("ts",))

    assert [(target.domain, target.language) for target in targets] == [
        ("packets", "ts"),
        ("constants", "ts"),
    ]


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

    constants_go = config.target("constants", "go")
    assert constants_go.files == (tmp_path / "services/game-server/internal/game/constants.go",)
    assert constants_go.sections == ("constants.gameplay", "constants.network")
    assert constants_go.owns == ("constants.gameplay", "constants.network")


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
