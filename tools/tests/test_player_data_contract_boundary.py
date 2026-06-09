from __future__ import annotations

import re
from pathlib import Path
import tomllib


REPO_ROOT = Path(__file__).resolve().parents[2]
STATS_TOML = REPO_ROOT / "shared" / "player_data" / "stats.toml"
MATCH_RESULT_TOML = REPO_ROOT / "shared" / "player_data" / "match_result.toml"
GO_TYPES = REPO_ROOT / "services" / "game-server" / "internal" / "playerdata" / "types.go"


def test_player_data_stats_fields_match_go_contract() -> None:
    stats = tomllib.loads(STATS_TOML.read_text(encoding="utf-8"))
    go_text = GO_TYPES.read_text(encoding="utf-8")

    assert stats["schema_name"] == "stats"
    assert stats["schema_version"] == "v1.1"
    assert list(stats["fields"].keys()) == [
        "total_score",
        "high_score",
        "ship_deaths",
        "games_played",
        "wins",
    ]

    assert re.search(r"\bTotalScore\s+int\b", go_text)
    assert re.search(r"\bHighScore\s+int\b", go_text)
    assert re.search(r"\bShipDeaths\s+int\b", go_text)
    assert re.search(r"\bGamesPlayed\s+int\b", go_text)
    assert re.search(r"\bWins\s+int\b", go_text)
    assert 'Wins is account/multiplayer-only for V1.1.' in go_text
    assert stats["fields"]["wins"].get("scope") in {"account_only", "multiplayer_only"}


def test_player_data_match_result_fields_match_go_contract() -> None:
    match_result = tomllib.loads(MATCH_RESULT_TOML.read_text(encoding="utf-8"))
    go_text = GO_TYPES.read_text(encoding="utf-8")

    assert match_result["schema_name"] == "match_result"
    assert match_result["schema_version"] == "v1.1"

    assert "MatchResultSummary" in go_text
    assert "PlayerMatchSummary" in go_text
    assert re.search(r"\bGamePlayerID\s+string\b", go_text)
    assert re.search(r"\bAccountID\s+string\b", go_text)
    assert re.search(r"\bLocalProfileID\s+string\b", go_text)
    assert re.search(r"\bScore\s+int\b", go_text)
    assert re.search(r"\bShipDeaths\s+int\b", go_text)
    assert re.search(r"\bWon\s+bool\b", go_text)
    assert "AccountUserID" not in go_text

    assert list(match_result["MatchResultSummary"]["fields"].keys()) == [
        "match_id",
        "mode",
        "resolved_at",
    ]
    assert list(match_result["PlayerMatchSummary"]["fields"].keys()) == [
        "game_player_id",
        "account_id",
        "local_profile_id",
        "score",
        "ship_deaths",
        "won",
    ]
    assert "account_user_id" not in match_result["PlayerMatchSummary"]["fields"]
