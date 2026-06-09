from __future__ import annotations

from pathlib import Path

from data_sync.config import DEFAULT_CONSTANTS_SCAN, DataSyncConfig, ScanConfig
from data_sync.discovery import discover_constants_files, language_for_path


def test_language_for_path_returns_go_for_go_files() -> None:
    assert language_for_path(Path("services/game-server/internal/game/constants.go")) == "go"


def test_language_for_path_returns_gds_for_godot_files() -> None:
    assert language_for_path(Path("client/scripts/world/world_sync.gd")) == "gds"


def test_language_for_path_returns_ts_for_typescript_files() -> None:
    assert language_for_path(Path("services/api-server/src/constants.ts")) == "ts"


def test_language_for_path_returns_none_for_unsupported_suffixes() -> None:
    assert language_for_path(Path("shared/constants/weapons.toml")) is None


def test_discover_constants_files_finds_included_go_and_gd_files(tmp_path: Path) -> None:
    go_path = tmp_path / "services/game-server/internal/game/constants.go"
    gd_path = tmp_path / "client/scripts/world/world_sync.gd"
    go_path.parent.mkdir(parents=True)
    gd_path.parent.mkdir(parents=True)
    go_path.write_text(
        """
// data-sync:start constants.server.damage
const Damage = 10
// data-sync:end constants.server.damage
""".strip(),
        encoding="utf-8",
    )
    gd_path.write_text(
        """
# data-sync:start constants.client.hud
const HUD_MARGIN := 12
# data-sync:end constants.client.hud
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=DEFAULT_CONSTANTS_SCAN,
    )

    discovered = discover_constants_files(config, ("go", "gds"))

    assert [(item.path, item.language, item.sections) for item in discovered] == [
        (gd_path, "gds", ("constants.client.hud",)),
        (go_path, "go", ("constants.server.damage",)),
    ]


def test_discover_constants_files_ignores_excluded_directories(tmp_path: Path) -> None:
    included_path = tmp_path / "services/game-server/internal/game/constants.go"
    excluded_path = tmp_path / "client/.godot/editor/cache.gd"
    included_path.parent.mkdir(parents=True)
    excluded_path.parent.mkdir(parents=True)
    included_path.write_text(
        """
// data-sync:start constants.server.damage
const Damage = 10
// data-sync:end constants.server.damage
""".strip(),
        encoding="utf-8",
    )
    excluded_path.write_text(
        """
# data-sync:start constants.client.hud
const HUD_MARGIN := 12
# data-sync:end constants.client.hud
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=ScanConfig(
            include=("services/**/*.go", "client/**/*.gd"),
            exclude=("**/.godot/**",),
        ),
    )

    discovered = discover_constants_files(config, ("go", "gds"))

    assert [item.path for item in discovered] == [included_path]


def test_discover_constants_files_sorts_final_paths_globally(tmp_path: Path) -> None:
    client_path = tmp_path / "client/scripts/world/world_sync.gd"
    server_path = tmp_path / "services/game-server/internal/game/constants.go"
    client_path.parent.mkdir(parents=True)
    server_path.parent.mkdir(parents=True)
    client_path.write_text(
        """
# data-sync:start constants.client.hud
const HUD_MARGIN := 12
# data-sync:end constants.client.hud
""".strip(),
        encoding="utf-8",
    )
    server_path.write_text(
        """
// data-sync:start constants.server.damage
const Damage = 10
// data-sync:end constants.server.damage
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=ScanConfig(
            include=("services/**/*.go", "client/**/*.gd"),
            exclude=(),
        ),
    )

    discovered = discover_constants_files(config, ("go", "gds"))

    assert [item.path for item in discovered] == [client_path, server_path]


def test_discover_constants_files_ignores_unsupported_extensions(tmp_path: Path) -> None:
    toml_path = tmp_path / "services/game-server/internal/game/constants.toml"
    toml_path.parent.mkdir(parents=True)
    toml_path.write_text("value = 1", encoding="utf-8")

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=ScanConfig(
            include=("services/**/*.toml",),
            exclude=(),
        ),
    )

    assert discover_constants_files(config, ("go", "gds", "ts")) == ()


def test_discover_constants_files_excludes_nested_godot_paths(tmp_path: Path) -> None:
    included_path = tmp_path / "client/scripts/world/world_sync.gd"
    excluded_path = tmp_path / "client/.godot/editor/cache.gd"
    included_path.parent.mkdir(parents=True)
    excluded_path.parent.mkdir(parents=True)
    included_path.write_text(
        """
# data-sync:start constants.client.hud
const HUD_MARGIN := 12
# data-sync:end constants.client.hud
""".strip(),
        encoding="utf-8",
    )
    excluded_path.write_text(
        """
# data-sync:start constants.client.hud
const HUD_MARGIN := 16
# data-sync:end constants.client.hud
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=ScanConfig(
            include=("client/**/*.gd",),
            exclude=("**/.godot/**",),
        ),
    )

    discovered = discover_constants_files(config, ("gds",))

    assert [item.path for item in discovered] == [included_path]


def test_discover_constants_files_ignores_packet_only_markers(tmp_path: Path) -> None:
    go_path = tmp_path / "services/game-server/internal/network/packets.go"
    go_path.parent.mkdir(parents=True)
    go_path.write_text(
        """
// data-sync:start packets
const PacketJoin = 100
// data-sync:end packets
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=ScanConfig(
            include=("services/**/*.go",),
            exclude=(),
        ),
    )

    assert discover_constants_files(config, ("go",)) == ()


def test_discover_constants_files_filters_requested_languages(tmp_path: Path) -> None:
    go_path = tmp_path / "services/game-server/internal/game/constants.go"
    gd_path = tmp_path / "client/scripts/world/world_sync.gd"
    go_path.parent.mkdir(parents=True)
    gd_path.parent.mkdir(parents=True)
    go_path.write_text(
        """
// data-sync:start constants.server.damage
const Damage = 10
// data-sync:end constants.server.damage
""".strip(),
        encoding="utf-8",
    )
    gd_path.write_text(
        """
# data-sync:start constants.client.hud
const HUD_MARGIN := 12
# data-sync:end constants.client.hud
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=DEFAULT_CONSTANTS_SCAN,
    )

    discovered = discover_constants_files(config, ("gds",))

    assert [(item.path, item.language, item.sections) for item in discovered] == [
        (gd_path, "gds", ("constants.client.hud",)),
    ]


def test_discover_constants_files_finds_real_style_generated_constants_files(tmp_path: Path) -> None:
    server_constants_path = tmp_path / "services/game-server/internal/constants/constants.go"
    weapons_constants_path = tmp_path / "services/game-server/internal/constants/weapons.go"
    client_constants_path = tmp_path / "client/scripts/generated/constants/constants.gd"
    server_constants_path.parent.mkdir(parents=True)
    client_constants_path.parent.mkdir(parents=True)

    server_constants_path.write_text(
        """
package constants

// data-sync:start constants.server.runtime
const TickRate = 60
// data-sync:end constants.server.runtime
""".strip(),
        encoding="utf-8",
    )
    weapons_constants_path.write_text(
        """
package constants

// data-sync:start constants.server.weapons.basic_cannon
const BasicCannonDamage = 1
// data-sync:end constants.server.weapons.basic_cannon
""".strip(),
        encoding="utf-8",
    )
    client_constants_path.write_text(
        """
extends RefCounted

# data-sync:start constants.client.presentation.background
const STARFIELD_DENSITY := 4
# data-sync:end constants.client.presentation.background
""".strip(),
        encoding="utf-8",
    )

    config = DataSyncConfig(
        path=tmp_path / "config.toml",
        root=tmp_path,
        sot_paths_by_domain={},
        targets_by_domain_language={},
        constants_scan=ScanConfig(
            include=("services/**/*.go", "client/**/*.gd", "services/**/*.ts"),
            exclude=(".git/**", "**/.godot/**", "**/node_modules/**"),
        ),
    )

    discovered = discover_constants_files(config, ("go", "gds"))

    assert [(item.path, item.sections) for item in discovered] == [
        (client_constants_path, ("constants.client.presentation.background",)),
        (server_constants_path, ("constants.server.runtime",)),
        (weapons_constants_path, ("constants.server.weapons.basic_cannon",)),
    ]
