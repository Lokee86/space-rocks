from __future__ import annotations

import urllib.request
from pathlib import Path
from typing import Any


class HeapProfileCollector:
    def __init__(
        self,
        rounds: list[int],
        base_url: str,
        run_directory: Path,
    ) -> None:
        self.rounds = sorted(set(rounds))
        self.base_url = base_url.rstrip("/")
        self.run_directory = run_directory
        self.captured: dict[int, Path] = {}

    def capture_available(self) -> None:
        completed = self._completed_round_count()
        for round_number in self.rounds:
            if round_number > completed or round_number in self.captured:
                continue
            path = self.run_directory / "heap-profiles" / f"heap-round-{round_number:03d}.pb.gz"
            path.parent.mkdir(parents=True, exist_ok=True)
            with urllib.request.urlopen(
                f"{self.base_url}/debug/pprof/heap?gc=1",
                timeout=20.0,
            ) as response:
                path.write_bytes(response.read())
            self.captured[round_number] = path

    def summary(self) -> list[dict[str, Any]]:
        return [
            {"round": round_number, "path": str(path)}
            for round_number, path in sorted(self.captured.items())
        ]

    def _completed_round_count(self) -> int:
        directory = self.run_directory / "measurements" / "coordinator-1"
        if not directory.exists():
            return 0
        return len(list(directory.glob("measurement-v1-*.json")))
