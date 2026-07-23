from __future__ import annotations

import os
from pathlib import Path
import platform
import plistlib
import shutil
import tempfile

from .local_alpha_common import ROOT, ReleaseGateError, run

CLIENT_DIR = ROOT / "client"
GAME_SERVER_DIR = ROOT / "services" / "game-server"
CREDENTIAL_HELPER_DIR = CLIENT_DIR / "native" / "credential-helper"
DEFAULT_OUTPUT_ROOT = ROOT / "dist" / "local-alpha"
COLLISION_SHAPES_SOURCE = ROOT / "shared" / "collisions" / "collision_shapes.json"


def native_architectures(platform_name: str) -> tuple[str, str]:
    if platform_name == "windows":
        return "x86_64", "amd64"
    machine = platform.machine().lower()
    if machine in {"arm64", "aarch64"}:
        return "arm64", "arm64"
    if machine in {"x86_64", "amd64"}:
        return "x86_64", "amd64"
    raise ReleaseGateError(f"unsupported {platform_name} architecture: {machine}")


def godot_binary(explicit: str | None) -> str:
    if explicit:
        return explicit
    if value := os.environ.get("GODOT_BIN"):
        return value
    if os.name == "nt" and Path(r"C:\Godot.exe").exists():
        return r"C:\Godot.exe"
    return "godot"


def platform_layout(platform_name: str, output_root: Path) -> tuple[Path, Path, Path, str]:
    if platform_name == "windows":
        output_dir = output_root / "windows"
        return (
            output_dir,
            output_dir / "SpaceRocks.exe",
            output_dir / "space-rocks-server.exe",
            "Windows Local Alpha",
        )
    output_dir = output_root / "macos"
    app = output_dir / "Space Rocks.app"
    return (
        output_dir,
        app / "Contents" / "MacOS" / "Space Rocks",
        app / "Contents" / "Helpers" / "space-rocks-server",
        "macOS Local Alpha",
    )


def credential_helper_path(platform_name: str, output_dir: Path) -> Path:
    if platform_name == "windows":
        return output_dir / "space-rocks-credential-helper.exe"
    return output_dir / "Space Rocks.app" / "Contents" / "Helpers" / "space-rocks-credential-helper"


def package_collision_shapes(server_output: Path) -> Path:
    destination = server_output.parent / "shared" / "collisions" / "collision_shapes.json"
    destination.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(COLLISION_SHAPES_SOURCE, destination)
    return destination


def client_export_output(platform_name: str, client_executable: Path) -> Path:
    if platform_name == "macos":
        return client_executable.parents[2]
    return client_executable


def resolve_client_executable(platform_name: str, expected_path: Path) -> Path:
    if platform_name != "macos":
        return expected_path

    info_plist = expected_path.parents[1] / "Info.plist"
    try:
        with info_plist.open("rb") as handle:
            bundle_metadata = plistlib.load(handle)
    except (OSError, plistlib.InvalidFileException) as error:
        raise ReleaseGateError(f"could not read macOS bundle metadata: {info_plist}") from error

    executable_name = bundle_metadata.get("CFBundleExecutable")
    if not isinstance(executable_name, str) or not executable_name:
        raise ReleaseGateError(f"macOS bundle metadata is missing CFBundleExecutable: {info_plist}")
    if Path(executable_name).name != executable_name:
        raise ReleaseGateError(f"macOS bundle has invalid CFBundleExecutable: {executable_name!r}")
    return expected_path.parent / executable_name


def build_client(platform_name: str, godot: str, preset: str, client_executable: Path) -> None:
    client_executable.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="space-rocks-export-") as temporary:
        staged_client = Path(temporary) / "client"
        shutil.copytree(CLIENT_DIR, staged_client, ignore=_ignore_transient_files)
        run(
            [
                godot,
                "--headless",
                "--path",
                staged_client,
                "--export-release",
                preset,
                client_export_output(platform_name, client_executable),
            ],
            cwd=ROOT,
            timeout=600,
        )


def _ignore_transient_files(directory: str, names: list[str]) -> set[str]:
    ignored = {name for name in names if name in {".godot", ".import", ".export"}}
    directory_path = Path(directory)
    if directory_path.name == "lib":
        ignored.update(name for name in names if name.startswith("~"))
    if directory_path.name == "credential-helper":
        ignored.add("bin")
    if directory_path.name == "addons":
        ignored.add("engineforge_bridge")
    return ignored


def build_server(platform_name: str, server_output: Path, version: str) -> None:
    _, go_architecture = native_architectures(platform_name)
    environment = os.environ.copy()
    environment.update(
        {
            "GOOS": "windows" if platform_name == "windows" else "darwin",
            "GOARCH": go_architecture,
            "CGO_ENABLED": "0",
        }
    )
    server_output.parent.mkdir(parents=True, exist_ok=True)
    run(
        [
            "go",
            "build",
            "-trimpath",
            "-tags",
            "localpackage",
            "-ldflags",
            f"-s -w -X main.packagedBuildVersion={version}",
            "-o",
            server_output,
            "./cmd/game-server",
        ],
        cwd=GAME_SERVER_DIR,
        env=environment,
        timeout=600,
    )


def build_credential_helper(platform_name: str, helper_output: Path) -> None:
    _, go_architecture = native_architectures(platform_name)
    environment = os.environ.copy()
    environment.update(
        {
            "GOOS": "windows" if platform_name == "windows" else "darwin",
            "GOARCH": go_architecture,
            "CGO_ENABLED": "0" if platform_name == "windows" else "1",
        }
    )
    helper_output.parent.mkdir(parents=True, exist_ok=True)
    run(
        ["go", "build", "-trimpath", "-ldflags", "-s -w", "-o", helper_output, "."],
        cwd=CREDENTIAL_HELPER_DIR,
        env=environment,
        timeout=300,
    )


def ad_hoc_sign_macos(output_dir: Path) -> None:
    app = output_dir / "Space Rocks.app"
    helpers = app / "Contents" / "Helpers"
    for helper in sorted(helpers.iterdir()):
        run(["codesign", "--force", "--sign", "-", helper], cwd=ROOT)
    run(["codesign", "--force", "--deep", "--sign", "-", app], cwd=ROOT)
    run(["codesign", "--verify", "--deep", "--strict", app], cwd=ROOT)
