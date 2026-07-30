#!/usr/bin/env python3
from pathlib import Path
from urllib.request import urlopen
import socket
import subprocess

root = Path(__file__).resolve().parent
env_file = root / ".env"
compose_file = root / "compose.yaml"
base = ["docker", "compose", "--env-file", str(env_file), "-f", str(compose_file)]
subprocess.run([*base, "ps"], cwd=root, check=True)
for url in ("http://127.0.0.1:8082/up", "http://127.0.0.1:8081/health"):
    with urlopen(url, timeout=10) as response:
        if response.status >= 400:
            raise SystemExit(f"health check failed: {url} -> {response.status}")
with socket.create_connection(("127.0.0.1", 8083), timeout=10):
    pass
print("Space Rocks hosted services are healthy.")
