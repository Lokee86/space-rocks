from __future__ import annotations

import sys
from pathlib import Path

TOOLS_ROOT = Path(__file__).resolve().parents[1]
if str(TOOLS_ROOT) not in sys.path:
    sys.path.insert(0, str(TOOLS_ROOT))

from runtime_scenarios.soak_summary import summarize_round_windows


def test_summarizes_head_to_tail_window_drift() -> None:
    rounds = [
        {
            "server": {
                "peak_rss_mib": 30 + index,
                "tick_average_us": 100 + index,
                "receiver_outbound_average_us": 500 + index,
            },
            "client": {
                "peak_memory_mib": 200 + index * 2,
                "frame_average_ms": 7 + index * 0.1,
            },
        }
        for index in range(6)
    ]

    result = summarize_round_windows(rounds, 2)

    assert result["server_peak_rss_mib"]["head_average"] == 30.5
    assert result["server_peak_rss_mib"]["tail_average"] == 34.5
    assert result["server_peak_rss_mib"]["tail_minus_head"] == 4.0
    assert result["client_peak_memory_mib"]["maximum"] == 210.0
