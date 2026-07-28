from __future__ import annotations

from pathlib import Path

from runtime_scenarios.heap_profiles import HeapProfileCollector


class FakeResponse:
    def __enter__(self) -> "FakeResponse":
        return self

    def __exit__(self, *_args: object) -> None:
        return None

    def read(self) -> bytes:
        return b"heap-profile"


def test_captures_configured_profile_after_round_completion(
    tmp_path: Path, monkeypatch
) -> None:
    run_directory = tmp_path / "run"
    measurements = run_directory / "measurements" / "coordinator-1"
    measurements.mkdir(parents=True)
    collector = HeapProfileCollector([1, 3], "http://127.0.0.1:9999", run_directory)
    requested: list[str] = []

    def fake_urlopen(url: str, timeout: float) -> FakeResponse:
        requested.append(f"{url}|{timeout}")
        return FakeResponse()

    monkeypatch.setattr("runtime_scenarios.heap_profiles.urllib.request.urlopen", fake_urlopen)
    collector.capture_available()
    assert collector.summary() == []

    (measurements / "measurement-v1-round-1.json").write_text("{}", encoding="utf-8")
    collector.capture_available()

    profile = run_directory / "heap-profiles" / "heap-round-001.pb.gz"
    assert profile.read_bytes() == b"heap-profile"
    assert requested == ["http://127.0.0.1:9999/debug/pprof/heap?gc=1|20.0"]
    assert collector.summary() == [{"round": 1, "path": str(profile)}]

    collector.capture_available()
    assert len(requested) == 1
