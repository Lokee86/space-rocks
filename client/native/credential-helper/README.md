# Credential Helper

This helper is the native secure-storage boundary for the Godot client.

- Windows encrypts the Space Rocks bearer token with user-scoped DPAPI and stores only the encrypted blob in the Godot user-data directory.
- macOS stores the token as a generic-password item in the current user's login Keychain.
- Requests and responses use one JSON object over standard input/output. Secrets are never passed through process arguments.
- Other platforms fail closed and do not persist account credentials.

Build on the target platform:

```bash
mkdir -p bin
case "$(go env GOOS)" in
  windows) go build -trimpath -ldflags="-s -w" -o bin/space-rocks-credential-helper.exe . ;;
  darwin) CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o bin/space-rocks-credential-helper . ;;
esac
```

`AuthCredentialStore` looks in `bin/` while the client runs in the editor. After building, verify the complete Godot-to-platform-store path from the repository root:

```bash
godot --headless --path client -s res://tools/verify_credential_helper.gd
```

Packaged placement:

- Windows: place `space-rocks-credential-helper.exe` beside the game executable.
- macOS: embed `space-rocks-credential-helper` at `Space Rocks.app/Contents/Helpers/space-rocks-credential-helper` and sign it with the application bundle.

The helper intentionally has no network access or Discord-specific behavior. It stores only the Space Rocks bearer credential issued by the API server.
