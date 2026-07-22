# Local Packaged Alpha Release Gate

Build and verify the release-shaped local single-player package on its native operating system.

Windows:

```bash
python tools/release/package_local_alpha.py --platform windows
```

macOS:

```bash
python3 tools/release/package_local_alpha.py --platform macos --adhoc-sign
```

The gate requires a clean Git worktree so the artifact manifest identifies the exact source commit. For temporary development diagnostics only, `--allow-dirty` permits packaging and records the dirty paths plus a `-dirty` version suffix.

The gate exports the Godot client, builds the `localpackage` game-server variant, builds the platform credential helper, assembles the package layout, and runs the exported client twice on an isolated loopback port. The first run executes a packaged DPAPI or Keychain credential round trip, creates and selects an isolated local profile, starts a real single-player session, completes the server-authoritative match with deterministic alpha devtools commands, validates the resolved match result, and waits for local statistics to be written. The second run proves the profile, default selection, games played, score, high score, and ship deaths survived a complete client/server restart, then cleans up the isolated profile. Each run also verifies that the client-owned server process exits.

Public release signing is intentionally separate from this development gate. Replace ad-hoc macOS signing and unsigned Windows artifacts with release certificates when credentials are available.
