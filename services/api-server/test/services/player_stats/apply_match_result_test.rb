require "test_helper"

class PlayerStats::ApplyMatchResultTest < ActiveSupport::TestCase
  setup do
    @user = User.create!(display_name: "Ada")
  end

  test "missing user returns invalid_input" do
    result = PlayerStats::ApplyMatchResult.call(
      user: nil,
      result_id: "result-1",
      match_id: "match-1",
      score: 0,
      ship_deaths: 0,
      won: false
    )

    assert_not result.success?
    assert_equal :invalid_input, result.error
  end

  test "missing result_id returns invalid_input" do
    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: nil,
      match_id: "match-1",
      score: 0,
      ship_deaths: 0,
      won: false
    )

    assert_not result.success?
    assert_equal :invalid_input, result.error
  end

  test "missing match_id returns invalid_input" do
    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-1",
      match_id: nil,
      score: 0,
      ship_deaths: 0,
      won: false
    )

    assert_not result.success?
    assert_equal :invalid_input, result.error
  end

  test "valid minimal input does not return invalid_input" do
    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-1",
      match_id: "match-1",
      score: 0,
      ship_deaths: 0,
      won: false
    )

    assert_predicate result, :success?
    assert_not_equal :invalid_input, result.error
  end

  test "first result creates player_stat" do
    assert_difference -> { PlayerStat.count }, 1 do
      result = PlayerStats::ApplyMatchResult.call(
        user: @user,
        result_id: "result-create-stat",
        match_id: "match-create-stat",
        score: 5,
        ship_deaths: 1,
        won: false
      )

      assert_predicate result, :success?
      assert_predicate result.player_stat, :present?
    end
  end

  test "games_played increments by 1" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-games-played",
      match_id: "match-games-played",
      score: 5,
      ship_deaths: 1,
      won: false
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-games-played-2",
      match_id: "match-games-played-2",
      score: 3,
      ship_deaths: 2,
      won: false
    )

    assert_equal 2, result.player_stat.games_played
  end

  test "total_score adds the result score" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-total-score",
      match_id: "match-total-score",
      score: 5,
      ship_deaths: 1,
      won: false
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-total-score-2",
      match_id: "match-total-score-2",
      score: 7,
      ship_deaths: 2,
      won: false
    )

    assert_equal 12, result.player_stat.total_score
  end

  test "high_score is set from first score" do
    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-high-score-first",
      match_id: "match-high-score-first",
      score: 8,
      ship_deaths: 1,
      won: false
    )

    assert_equal 8, result.player_stat.high_score
  end

  test "high_score increases when later score is higher" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-high-score-1",
      match_id: "match-high-score-1",
      score: 4,
      ship_deaths: 1,
      won: false
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-high-score-2",
      match_id: "match-high-score-2",
      score: 9,
      ship_deaths: 1,
      won: false
    )

    assert_equal 9, result.player_stat.high_score
  end

  test "high_score does not decrease when later score is lower" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-high-score-low-1",
      match_id: "match-high-score-low-1",
      score: 10,
      ship_deaths: 1,
      won: false
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-high-score-low-2",
      match_id: "match-high-score-low-2",
      score: 2,
      ship_deaths: 1,
      won: false
    )

    assert_equal 10, result.player_stat.high_score
  end

  test "ship_deaths add to existing deaths" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-ship-deaths-1",
      match_id: "match-ship-deaths-1",
      score: 5,
      ship_deaths: 2,
      won: false
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-ship-deaths-2",
      match_id: "match-ship-deaths-2",
      score: 3,
      ship_deaths: 4,
      won: false
    )

    assert_equal 6, result.player_stat.ship_deaths
  end

  test "wins increments when won is true" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-wins-1",
      match_id: "match-wins-1",
      score: 5,
      ship_deaths: 1,
      won: false
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-wins-2",
      match_id: "match-wins-2",
      score: 7,
      ship_deaths: 2,
      won: true
    )

    assert_equal 1, result.player_stat.wins
  end

  test "wins does not increment when won is false" do
    PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-wins-false-1",
      match_id: "match-wins-false-1",
      score: 5,
      ship_deaths: 1,
      won: true
    )

    result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-wins-false-2",
      match_id: "match-wins-false-2",
      score: 7,
      ship_deaths: 2,
      won: false
    )

    assert_equal 1, result.player_stat.wins
  end

  test "PlayerMatchResult row is recorded" do
    assert_difference -> { PlayerMatchResult.count }, 1 do
      result = PlayerStats::ApplyMatchResult.call(
        user: @user,
        result_id: "result-recorded",
        match_id: "match-recorded",
        score: 5,
        ship_deaths: 1,
        won: true
      )

      assert_predicate result.match_result, :present?
      assert_equal "result-recorded", result.match_result.result_id
    end
  end

  test "applying the same result_id twice is idempotent" do
    first_result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-idempotent",
      match_id: "match-idempotent",
      score: 6,
      ship_deaths: 2,
      won: true
    )

    second_result = PlayerStats::ApplyMatchResult.call(
      user: @user,
      result_id: "result-idempotent",
      match_id: "match-idempotent",
      score: 6,
      ship_deaths: 2,
      won: true
    )

    assert_predicate first_result, :success?
    assert_predicate second_result, :success?
    assert_equal true, second_result.duplicate
    assert_equal first_result.player_stat.id, second_result.player_stat.id
    assert_equal first_result.match_result.id, second_result.match_result.id
    assert_equal 1, second_result.player_stat.games_played
    assert_equal 6, second_result.player_stat.total_score
    assert_equal 2, second_result.player_stat.ship_deaths
    assert_equal 1, second_result.player_stat.wins
    assert_equal 1, PlayerMatchResult.where(result_id: "result-idempotent").count
  end
end
