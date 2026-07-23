import plistlib
from pathlib import Path

import pytest

from tools.release.local_alpha_build import (
    CLIENT_DIR,
    client_export_output,
    credential_helper_path,
    native_architectures,
    package_collision_shapes,
    platform_layout,
    resolve_client_executable,
)
from tools.release.local_alpha_common import ReleaseGateError
from tools.release.local_alpha_manifest import default_version, package_files
from tools.release.local_alpha_smoke import (
    CREDENTIAL_KEYCHAIN_PATH_ENVIRONMENT,
    prepare_macos_smoke_keychain,
    remove_macos_smoke_keychain,
    smoke_environment,
)


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


def test_collision_shapes_are_packaged_beside_the_server(tmp_path: Path, monkeypatch) -> None:
    source = tmp_path / "source" / "collision_shapes.json"
    source.parent.mkdir()
    source.write_text('{"ship": {"type": "polygon"}}', encoding="utf-8")
    monkeypatch.setattr("tools.release.local_alpha_build.COLLISION_SHAPES_SOURCE", source)
    server = tmp_path / "package" / "space-rocks-server.exe"

    destination = package_collision_shapes(server)

    assert destination == server.parent / "shared" / "collisions" / "collision_shapes.json"
    assert destination.read_text(encoding="utf-8") == source.read_text(encoding="utf-8")


def test_macos_client_executable_comes_from_bundle_metadata(tmp_path: Path) -> None:
    expected = tmp_path / "Space Rocks.app" / "Contents" / "MacOS" / "Space Rocks"
    info_plist = expected.parents[1] / "Info.plist"
    info_plist.parent.mkdir(parents=True)
    with info_plist.open("wb") as handle:
        plistlib.dump({"CFBundleExecutable": "SpaceRocks"}, handle)

    assert resolve_client_executable("macos", expected) == expected.parent / "SpaceRocks"


def test_macos_client_executable_rejects_invalid_bundle_metadata(tmp_path: Path) -> None:
    expected = tmp_path / "Space Rocks.app" / "Contents" / "MacOS" / "Space Rocks"
    info_plist = expected.parents[1] / "Info.plist"
    info_plist.parent.mkdir(parents=True)
    with info_plist.open("wb") as handle:
        plistlib.dump({"CFBundleExecutable": "../SpaceRocks"}, handle)

    with pytest.raises(ReleaseGateError, match="invalid CFBundleExecutable"):
        resolve_client_executable("macos", expected)


def test_macos_smoke_keychain_uses_the_isolated_smoke_home(
    tmp_path: Path,
    monkeypatch,
) -> None:
    calls = []
    environment = smoke_environment("macos", tmp_path, 43128)
    monkeypatch.setattr(
        "tools.release.local_alpha_smoke.secrets.token_urlsafe",
        lambda _length: "test-password",
    )

    def record_run(command, **kwargs):
        calls.append((command, kwargs))

    monkeypatch.setattr("tools.release.local_alpha_smoke.subprocess.run", record_run)

    keychain_path = prepare_macos_smoke_keychain(tmp_path, environment)

    assert keychain_path == tmp_path / "local-alpha-smoke.keychain-db"
    assert environment[CREDENTIAL_KEYCHAIN_PATH_ENVIRONMENT] == str(keychain_path)
    assert calls[0][0][:3] == ["security", "create-keychain", "-p"]
    assert calls[2][0][:3] == ["security", "unlock-keychain", "-p"]
    assert calls[3][0][:5] == ["security", "default-keychain", "-d", "user", "-s"]
    assert calls[4][0][:5] == ["security", "list-keychains", "-d", "user", "-s"]
    assert all(call[1]["env"] is environment for call in calls)
    assert all(call[1]["cwd"] == tmp_path for call in calls)

    remove_macos_smoke_keychain(keychain_path, environment)

    assert calls[5][0] == ["security", "delete-keychain", str(keychain_path)]
    assert CREDENTIAL_KEYCHAIN_PATH_ENVIRONMENT not in environment


def test_macos_export_uses_official_universal_template_configuration() -> None:
    project_settings = (CLIENT_DIR / "project.godot").read_text(encoding="utf-8")
    export_presets = (CLIENT_DIR / "export_presets.cfg").read_text(encoding="utf-8")

    assert "textures/vram_compression/import_etc2_astc=true" in project_settings
    assert "texture_format/etc2_astc=true" in export_presets
    assert 'binary_format/architecture="universal"' in export_presets


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
