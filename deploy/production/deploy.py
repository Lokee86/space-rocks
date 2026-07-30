#!/usr/bin/env python3
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parent
env_file = root / ".env"
compose_file = root / "compose.yaml"

if not env_file.is_file():
    raise SystemExit("Missing .env. Create it from the README template and fill every secret first.")
if "replace-with-" in env_file.read_text(encoding="utf-8"):
    raise SystemExit("Refusing to deploy while placeholder values remain in .env.")

base = ["docker", "compose", "--env-file", str(env_file), "-f", str(compose_file)]
for args in (["config", "--quiet"], ["pull"], ["up", "-d", "--remove-orphans"], ["ps"]):
    subprocess.run([*base, *args], cwd=root, check=True)
