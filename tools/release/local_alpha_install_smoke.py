from __future__ import annotations

from pathlib import Path
import shutil
import tempfile

from .local_alpha_common import ReleaseGateError, run


def run_installer_smoke(platform_name: str, output_dir: Path) -> None:
    with tempfile.TemporaryDirectory(prefix="space-rocks-install-") as temporary:
        temporary_root = Path(temporary)
        if platform_name == "windows":
            installed_root = _install_windows(output_dir, temporary_root)
            required = (
                installed_root / "SpaceRocks.exe",
                installed_root / "SpaceRocks.pck",
                installed_root / "space-rocks-server.exe",
                installed_root / "space-rocks-credential-helper.exe",
            )
        else:
            installed_root = _install_macos(output_dir, temporary_root)
            required = (
                installed_root / "Contents" / "Info.plist",
                installed_root / "Contents" / "Helpers" / "space-rocks-server",
                installed_root / "Contents" / "Helpers" / "space-rocks-credential-helper",
            )

        missing = [path for path in required if not path.exists()]
        if missing:
            raise ReleaseGateError(f"installer smoke test is missing installed files: {missing}")


def _install_windows(output_dir: Path, temporary_root: Path) -> Path:
    powershell = shutil.which("powershell.exe") or shutil.which("pwsh.exe")
    if powershell is None:
        raise ReleaseGateError("PowerShell is required to verify the Windows installer")
    installed_root = temporary_root / "Space Rocks"
    run(
        [
            powershell,
            "-NoProfile",
            "-ExecutionPolicy",
            "Bypass",
            "-File",
            output_dir / "install.ps1",
            "-InstallDir",
            installed_root,
            "-NoStartMenuShortcut",
        ],
        cwd=output_dir,
    )
    return installed_root


def _install_macos(output_dir: Path, temporary_root: Path) -> Path:
    applications = temporary_root / "Applications"
    run(
        ["/bin/sh", output_dir / "install.command", applications],
        cwd=output_dir,
    )
    return applications / "Space Rocks.app"
