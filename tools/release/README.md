# Local Packaged Alpha Release Gate

Build and verify the release-shaped local single-player package on its native operating system.

The current packaged release version is stored in `tools/release/version.txt`. Bump that file before creating the matching `v<version>` tag.

Windows:

```bash
python tools/release/package_local_alpha.py --platform windows
```

macOS:

```bash
python3 tools/release/package_local_alpha.py --platform macos --adhoc-sign
```

The gate requires a clean Git worktree so the artifact manifest identifies the exact source commit. For temporary development diagnostics only, `--allow-dirty` permits packaging and records the dirty paths plus a `-dirty` version suffix.

The gate exports the Godot client, builds the `localpackage` game-server variant, builds the platform credential helper, assembles the package layout, adds and smoke-tests the platform installer in a temporary location, and runs the exported client twice on an isolated loopback port. The first run executes a packaged DPAPI or Keychain credential round trip, creates and selects an isolated local profile, starts a real single-player session, completes the server-authoritative match with deterministic alpha devtools commands, validates the resolved match result, and waits for local statistics to be written. The second run proves the profile, default selection, games played, score, high score, and ship deaths survived a complete client/server restart, then cleans up the isolated profile. Each run also verifies that the client-owned server process exits.

## Installing a packaged release

Windows packages include `install.ps1`. Run it from an extracted package:

```powershell
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

It installs for the current user under `%LOCALAPPDATA%\Programs\Space Rocks` and creates a Start Menu shortcut. Use `-InstallDir PATH`, `-NoStartMenuShortcut`, or `-DesktopShortcut` to change that behavior.

macOS packages include `install.command`. Run it from an extracted package:

```bash
./install.command
```

It installs to `~/Applications/Space Rocks.app` by default. Pass another application directory as the first argument or set `SPACE_ROCKS_INSTALL_DIR`. An existing app is moved to a timestamped backup before the new package is copied.

## Publishing a version

1. Update `tools/release/version.txt`.
2. Commit the release changes to `main`.
3. Create and push the matching tag, for example `v0.1.1`.
4. The native Windows and macOS jobs build and smoke-test their packages.
5. The tag workflow publishes both tested archives as a GitHub prerelease.

Public release signing is intentionally separate from this development gate. Replace ad-hoc macOS signing and unsigned Windows artifacts with release certificates when credentials are available.
