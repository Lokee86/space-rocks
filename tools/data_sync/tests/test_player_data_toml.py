from __future__ import annotations

from pathlib import Path

import pytest

from data_sync.player_data_toml import (
    PlayerDataField,
    PlayerDataGroup,
    PlayerDataTomlError,
    load_player_data_schema,
)


def write_toml(path: Path, body: str) -> Path:
    path.write_text(body.strip() + "\n", encoding="utf-8")
    return path


def test_load_player_data_schema_parses_stats_file(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "stats.toml",
        """
schema_name = "stats"
schema_version = "v1.1"

[fields.total_score]
type = "integer"
default = 0

[fields.high_score]
type = "integer"
default = 0

[fields.ship_deaths]
type = "integer"
default = 0

[fields.games_played]
type = "integer"
default = 0

[fields.wins]
type = "integer"
default = 0
scope = "multiplayer_only"
""",
    )

    schema = load_player_data_schema(path)

    assert schema.name == "stats"
    assert schema.version == "v1.1"
    assert schema.groups == (
        PlayerDataGroup(
            name="stats",
            fields=(
                PlayerDataField(name="total_score", type="integer", required=False, default=0),
                PlayerDataField(name="high_score", type="integer", required=False, default=0),
                PlayerDataField(name="ship_deaths", type="integer", required=False, default=0),
                PlayerDataField(name="games_played", type="integer", required=False, default=0),
                PlayerDataField(name="wins", type="integer", required=False, default=0),
            ),
        ),
    )


def test_load_player_data_schema_parses_match_result_file(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "match_result.toml",
        """
schema_name = "match_result"
schema_version = "v1.1"

[MatchResultSummary.metadata]
winner_rule = "multiplayer_highest_score"
ties_award_no_wins = true

[MatchResultSummary.fields.match_id]
type = "string"
required = true

[MatchResultSummary.fields.mode]
type = "string"
required = true

[MatchResultSummary.fields.resolved_at]
type = "string"
optional = true

[PlayerMatchSummary.fields.game_player_id]
type = "string"
required = true

[PlayerMatchSummary.fields.account_user_id]
type = "integer"
optional = true

[PlayerMatchSummary.fields.local_profile_id]
type = "string"
optional = true

[PlayerMatchSummary.fields.score]
type = "integer"
default = 0

[PlayerMatchSummary.fields.ship_deaths]
type = "integer"
default = 0

[PlayerMatchSummary.fields.won]
type = "boolean"
default = false
""",
    )

    schema = load_player_data_schema(path)

    assert schema.name == "match_result"
    assert schema.version == "v1.1"
    assert [group.name for group in schema.groups] == ["MatchResultSummary", "PlayerMatchSummary"]
    assert schema.groups[0].fields[0].name == "match_id"
    assert schema.groups[0].fields[0].type == "string"
    assert schema.groups[0].fields[0].required is True
    assert schema.groups[0].fields[2].name == "resolved_at"
    assert schema.groups[0].fields[2].required is False
    assert schema.groups[1].fields[0].name == "game_player_id"
    assert schema.groups[1].fields[1].name == "account_user_id"
    assert schema.groups[1].fields[1].required is False
    assert schema.groups[1].fields[3].name == "score"
    assert schema.groups[1].fields[3].default == 0
    assert schema.groups[1].fields[5].name == "won"
    assert schema.groups[1].fields[5].type == "boolean"


def test_load_player_data_schema_rejects_missing_field_name(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "missing_field_name.toml",
        """
schema_name = "stats"
schema_version = "v1.1"

[fields.""]
type = "integer"
default = 0
""",
    )

    with pytest.raises(PlayerDataTomlError, match=r"missing name"):
        load_player_data_schema(path)


def test_load_player_data_schema_rejects_unsupported_field_type(tmp_path: Path) -> None:
    path = write_toml(
        tmp_path / "unsupported_type.toml",
        """
schema_name = "stats"
schema_version = "v1.1"

[fields.total_score]
type = "float"
default = 0
""",
    )

    with pytest.raises(PlayerDataTomlError, match=r"must be one of: string, integer, boolean"):
        load_player_data_schema(path)
