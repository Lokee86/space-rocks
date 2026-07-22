from __future__ import annotations

import os
from pathlib import Path
import subprocess
from typing import Sequence

ROOT = Path(__file__).resolve().parents[2]


class ReleaseGateError(RuntimeError):
    pass


def run(
    command: Sequence[str | os.PathLike[str]],
    *,
    cwd: Path,
    env: dict[str, str] | None = None,
    timeout: int = 300,
) -> None:
    printable = " ".join(str(part) for part in command)
    print(f"+ {printable}", flush=True)
    subprocess.run(
        [str(part) for part in command],
        cwd=cwd,
        env=env,
        check=True,
        timeout=timeout,
    )
