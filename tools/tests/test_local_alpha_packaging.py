from pathlib import Path

from tools.release.local_alpha_build import (
    CLIENT_DIR,
    client_export_output,
    credential_helper_path,
    native_architectures,
    platform_layout,
)
from tools.release.local_alpha_manifest import default_version, package_files
from tools.release.local_alpha_smoke import smoke_environment


def test_dirty_default_version_is_visibly_marked(monkeypatch) -> None:
    monkeypatch.setattr("tools.release.local_alpha_manifest.git_commit", lambda: "abc123")

    assert default_version() == "local-alpha-abc123"
    assert default_version(dirty=True) == "local-alpha-abc123-dirty"


def test_native_architecture_mapping(monkeypatch) -> None:
    assert native_architectures("windows") == ("x86_64", "amd64")

    monkeypatch.setattr("tools.release.local_alpha_build.platform.machine", lambda: "arm64")
    assert native_architectures("macos") == ("arm64", "arm64")

    monkeypatch.setattr("tools.release.local_alpha_build.platform.machine", lambda: "x86_64")
    assert native_architectures("macos") == ("x86_64", "amd64")


def test_windows_package_layout() -> None:
    output = Path("dist/local-alpha")
    output_dir, client, server, preset = platform_layout("windows", output)

    assert output_dir == output / "windows"
    assert client == output / "windows" / "SpaceRocks.exe"
    assert server == output / "windows" / "space-rocks-server.exe"
    assert credential_helper_path("windows", output_dir) == output_dir / "space-rocks-credential-helper.exe"
    assert preset == "Windows Local Alpha"


def test_macos_package_layout() -> None:
    output = Path("dist/local-alpha")
    output_dir, client, server, preset = platform_layout("macos", output)

    app = output / "macos" / "Space Rocks.app"
    assert client == app / "Contents" / "MacOS" / "Space Rocks"
    assert server == app / "Contents" / "Helpers" / "space-rocks-server"
    assert credential_helper_path("macos", output_dir) == app / "Contents" / "Helpers" / "space-rocks-credential-helper"
    assert preset == "macOS Local Alpha"
    assert client_export_output("macos", client) == app


def test_windows_export_targets_the_executable() -> None:
    executable = Path("dist/local-alpha/windows/SpaceRocks.exe")

    assert client_export_output("windows", executable) == executable


def test_macos_export_texture_compression_is_enabled() -> None:
    project_settings = (CLIENT_DIR / "project.godot").read_text(encoding="utf-8")
    export_presets = (CLIENT_DIR / "export_presets.cfg").read_text(encoding="utf-8")

    assert "textures/vram_compression/import_etc2_astc=true" in project_settings
    assert "texture_format/etc2_astc=true" in export_presets


def test_manifest_file_list_includes_the_complete_package(tmp_path: Path) -> None:
    (tmp_path / "SpaceRocks.exe").write_bytes(b"client")
    (tmp_path / "SpaceRocks.pck").write_bytes(b"pck")
    helpers = tmp_path / "nested"
    helpers.mkdir()
    (helpers / "dependency.dll").write_bytes(b"dll")
    (tmp_path / "release-manifest.json").write_text("{}", encoding="utf-8")

    assert package_files(tmp_path) == sorted(
        [
            tmp_path / "SpaceRocks.exe",
            tmp_path / "SpaceRocks.pck",
            helpers / "dependency.dll",
        ]
    )


def test_smoke_environment_is_isolated(tmp_path: Path) -> None:
    windows = smoke_environment("windows", tmp_path, 43127)
    macos = smoke_environment("macos", tmp_path, 43128)

    assert windows["SPACE_ROCKS_LOCAL_SERVER_PORT"] == "43127"
    assert macos["SPACE_ROCKS_LOCAL_SERVER_PORT"] == "43128"
    assert windows["APPDATA"].startswith(str(tmp_path))
    assert windows["LOCALAPPDATA"].startswith(str(tmp_path))
    assert macos["HOME"] == str(tmp_path)
    assert macos["XDG_CONFIG_HOME"].startswith(str(tmp_path))
