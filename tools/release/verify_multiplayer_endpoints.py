from __future__ import annotations

from pathlib import Path
import sys

EXPECTED = (
    b"wss://game.laughingskull.ca/ws",
    b"https://api.laughingskull.ca",
    b"https://game.laughingskull.ca",
)
FORBIDDEN = (
    b"game.space-rocks.laughingskull.ca",
    b"space-rocks.laughingskull.ca/game",
)
SOURCE_FILES = (
    Path("client/scripts/boot/session_network_target.gd"),
    Path("client/scripts/api/api_config.gd"),
)
PRESET_REQUIREMENTS = (
    b'name="Windows Multiplayer Alpha"',
    b'custom_features="multiplayer_alpha"',
    b'export_path="../dist/multiplayer-alpha/windows/SpaceRocks.exe"',
    b'debug/export_console_wrapper=0',
    b'name="macOS Multiplayer Alpha"',
    b'export_path="../dist/multiplayer-alpha/macos/Space Rocks.app"',
)


def main() -> int:
    root = Path.cwd()
    source = b"\n".join((root / path).read_bytes() for path in SOURCE_FILES)
    missing = [value.decode() for value in EXPECTED if value not in source]
    forbidden = [value.decode() for value in FORBIDDEN if value in source]

    presets = (root / "client/export_presets.cfg").read_bytes()
    missing_presets = [value.decode() for value in PRESET_REQUIREMENTS if value not in presets]

    if missing or forbidden or missing_presets:
        details = []
        if missing:
            details.append("missing endpoints: " + ", ".join(missing))
        if forbidden:
            details.append("forbidden endpoints: " + ", ".join(forbidden))
        if missing_presets:
            details.append("missing preset settings: " + ", ".join(missing_presets))
        raise RuntimeError("multiplayer release verification failed (" + "; ".join(details) + ")")

    print("verified Windows and macOS multiplayer release endpoints and export paths")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, RuntimeError) as exc:
        print(exc, file=sys.stderr)
        raise SystemExit(1)
