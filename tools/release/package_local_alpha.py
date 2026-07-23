#!/usr/bin/env python3
"""Build and verify the local packaged single-player alpha artifact."""

from __future__ import annotations

import argparse
import os
from pathlib import Path
import shutil
import subprocess
import sys

sys.dont_write_bytecode = True

if __package__ in {None, ""}:
    sys.path.insert(0, str(Path(__file__).resolve().parents[2]))

from tools.release.local_alpha_build import (  # noqa: E402
    DEFAULT_OUTPUT_ROOT,
    ad_hoc_sign_macos,
    build_client,
    build_credential_helper,
    build_server,
    credential_helper_path,
    godot_binary,
    platform_layout,
    resolve_client_executable,
)
from tools.release.local_alpha_common import ReleaseGateError  # noqa: E402
from tools.release.local_alpha_manifest import (  # noqa: E402
    default_version,
    git_worktree_changes,
    package_files,
    write_manifest,
)
from tools.release.local_alpha_smoke import run_packaged_smoke  # noqa: E402


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--platform", choices=("windows", "macos"), required=True)
    parser.add_argument("--godot")
    parser.add_argument("--version")
    parser.add_argument("--output-root", type=Path, default=DEFAULT_OUTPUT_ROOT)
    parser.add_argument("--skip-smoke", action="store_true")
    parser.add_argument("--keep-output", action="store_true")
    parser.add_argument("--adhoc-sign", action="store_true", help="Ad-hoc sign the macOS app and helpers")
    parser.add_argument(
        "--allow-dirty",
        action="store_true",
        help="Permit a development-only package from uncommitted source and mark its manifest dirty",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    host_platform = "windows" if os.name == "nt" else "macos" if sys.platform == "darwin" else "other"
    if host_platform != args.platform:
        raise ReleaseGateError(f"{args.platform} packages must be built on a native {args.platform} runner")

    worktree_changes = git_worktree_changes()
    if worktree_changes and not args.allow_dirty:
        preview = ", ".join(worktree_changes[:5])
        raise ReleaseGateError(
            "release packages require a clean Git worktree; commit or restore changes first "
            f"(found: {preview})"
        )

    version = args.version or default_version(dirty=bool(worktree_changes))
    output_root = args.output_root.resolve()
    output_dir, client_executable, server_output, preset = platform_layout(args.platform, output_root)
    helper_output = credential_helper_path(args.platform, output_dir)

    if output_dir.exists() and not args.keep_output:
        shutil.rmtree(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    build_client(args.platform, godot_binary(args.godot), preset, client_executable)
    client_executable = resolve_client_executable(args.platform, client_executable)
    build_server(args.platform, server_output, version)
    build_credential_helper(args.platform, helper_output)

    if args.platform == "macos":
        server_output.chmod(0o755)
        helper_output.chmod(0o755)
        if args.adhoc_sign:
            ad_hoc_sign_macos(output_dir)

    required_files = [client_executable, server_output, helper_output]
    missing = [path for path in required_files if not path.is_file()]
    if missing:
        raise ReleaseGateError(f"package is missing required files: {missing}")

    if not args.skip_smoke:
        run_packaged_smoke(args.platform, client_executable)

    write_manifest(
        args.platform,
        output_dir,
        version,
        package_files(output_dir),
        worktree_changes,
    )
    print(f"local packaged alpha automated gate passed: {output_dir}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (ReleaseGateError, subprocess.CalledProcessError, subprocess.TimeoutExpired) as error:
        print(f"release gate failed: {error}", file=sys.stderr)
        raise SystemExit(1)
