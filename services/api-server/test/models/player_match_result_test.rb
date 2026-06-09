require "test_helper"

class PlayerMatchResultTest < ActiveSupport::TestCase
  test "valid result belongs to a user" do
    user = User.create!(display_name: "Ada")
    result = PlayerMatchResult.create!(
      user: user,
      result_id: "result-1",
      match_id: "match-1",
      score: 12,
      ship_deaths: 3,
      won: true
    )

    assert_equal user, result.user
  end

  test "result_id is required" do
    result = build_result(result_id: nil)

    assert_not result.valid?
    assert_includes result.errors[:result_id], "can't be blank"
  end

  test "match_id is required" do
    result = build_result(match_id: nil)

    assert_not result.valid?
    assert_includes result.errors[:match_id], "can't be blank"
  end

  test "result_id must be unique" do
    PlayerMatchResult.create!(
      user: User.create!(display_name: "Ada"),
      result_id: "result-1",
      match_id: "match-1",
      score: 12,
      ship_deaths: 3,
      won: true
    )

    duplicate = build_result(result_id: "result-1")

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:result_id], "has already been taken"
  end

  test "negative score is invalid" do
    result = build_result(score: -1)

    assert_not result.valid?
    assert_includes result.errors[:score], "must be greater than or equal to 0"
  end

  test "negative ship_deaths is invalid" do
    result = build_result(ship_deaths: -1)

    assert_not result.valid?
    assert_includes result.errors[:ship_deaths], "must be greater than or equal to 0"
  end

  test "won must be boolean" do
    result = build_result(won: nil)

    assert_not result.valid?
    assert_includes result.errors[:won], "is not included in the list"
  end

  test "user association exists" do
    user = User.create!(display_name: "Ada")
    result = PlayerMatchResult.create!(
      user: user,
      result_id: "result-2",
      match_id: "match-2",
      score: 0,
      ship_deaths: 0,
      won: false
    )

    assert_equal user, result.user
  end

  private

  def build_result(attributes = {})
    PlayerMatchResult.new(
      {
        user: User.create!(display_name: "Ada"),
        result_id: "result-default",
        match_id: "match-default",
        score: 0,
        ship_deaths: 0,
        won: true
      }.merge(attributes)
    )
  end
end
