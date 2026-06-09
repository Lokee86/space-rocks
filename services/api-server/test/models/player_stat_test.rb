require "test_helper"

class PlayerStatTest < ActiveSupport::TestCase
  test "a user can have one player_stat" do
    user = User.create!(display_name: "One")

    player_stat = PlayerStat.create!(user: user)

    assert_equal player_stat, user.reload.player_stat
  end

  test "default values are zero after creation" do
    player_stat = PlayerStat.create!(user: User.create!(display_name: "One"))

    assert_equal 0, player_stat.total_score
    assert_equal 0, player_stat.high_score
    assert_equal 0, player_stat.ship_deaths
    assert_equal 0, player_stat.games_played
    assert_equal 0, player_stat.wins
  end

  test "user_id uniqueness is enforced" do
    user = User.create!(display_name: "One")
    PlayerStat.create!(user: user)

    duplicate = PlayerStat.new(user: user)

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:user_id], "has already been taken"
  end

  test "negative total_score is invalid" do
    player_stat = build_player_stat(total_score: -1)

    assert_not player_stat.valid?
    assert_includes player_stat.errors[:total_score], "must be greater than or equal to 0"
  end

  test "negative high_score is invalid" do
    player_stat = build_player_stat(high_score: -1)

    assert_not player_stat.valid?
    assert_includes player_stat.errors[:high_score], "must be greater than or equal to 0"
  end

  test "negative ship_deaths is invalid" do
    player_stat = build_player_stat(ship_deaths: -1)

    assert_not player_stat.valid?
    assert_includes player_stat.errors[:ship_deaths], "must be greater than or equal to 0"
  end

  test "negative games_played is invalid" do
    player_stat = build_player_stat(games_played: -1)

    assert_not player_stat.valid?
    assert_includes player_stat.errors[:games_played], "must be greater than or equal to 0"
  end

  test "negative wins is invalid" do
    player_stat = build_player_stat(wins: -1)

    assert_not player_stat.valid?
    assert_includes player_stat.errors[:wins], "must be greater than or equal to 0"
  end

  private

  def build_player_stat(attributes = {})
    PlayerStat.new(
      {
        user: User.create!(display_name: "One"),
        total_score: 0,
        high_score: 0,
        ship_deaths: 0,
        games_played: 0,
        wins: 0
      }.merge(attributes)
    )
  end
end
