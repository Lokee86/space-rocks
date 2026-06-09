require "test_helper"

class Internal::PlayerData::MatchResultsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = User.create!(display_name: "Ada")
  end

  test "POST /internal/player-data/match-results without an Authorization header returns 401" do
    with_internal_token_env do
      post "/internal/player-data/match-results"
    end

    assert_response :unauthorized
  end

  test "POST /internal/player-data/match-results with a malformed Authorization header returns 401" do
    with_internal_token_env do
      post "/internal/player-data/match-results", headers: { "Authorization" => "Token test-internal-token" }
    end

    assert_response :unauthorized
  end

  test "POST /internal/player-data/match-results with the wrong bearer token returns 401" do
    with_internal_token_env do
      post "/internal/player-data/match-results", headers: internal_headers("wrong-token")
    end

    assert_response :unauthorized
  end

  test "POST /internal/player-data/match-results with a valid internal secret and missing required params returns 422" do
    with_internal_token_env do
      post "/internal/player-data/match-results", headers: internal_headers
    end

    assert_response :unprocessable_entity
    assert_equal({ "accepted" => false, "error" => "invalid_input" }, JSON.parse(response.body))
  end

  test "POST /internal/player-data/match-results with a valid internal secret and valid payload returns success" do
    with_internal_token_env do
      post "/internal/player-data/match-results",
        params: {
          result_id: "result-1",
          match_id: "match-1",
          account_id: @user.account_id,
          score: 12,
          ship_deaths: 3,
          won: true
        },
        headers: internal_headers
    end

    assert_response :success

    body = JSON.parse(response.body)

    assert_equal true, body["accepted"]
    assert_equal false, body["duplicate"]
    assert_equal 1, body["stats"]["games_played"]
    assert_equal 12, body["stats"]["total_score"]
    assert_equal 12, body["stats"]["high_score"]
    assert_equal 3, body["stats"]["ship_deaths"]
    assert_equal 1, body["stats"]["wins"]
    refute_includes response.body, "email"
    refute_includes response.body, "token_digest"
    refute_includes response.body, "password_digest"
    refute_includes response.body, "access_token"
  end

  test "POST /internal/player-data/match-results with an unknown account_id returns 404" do
    with_internal_token_env do
      post "/internal/player-data/match-results",
        params: {
          result_id: "result-unknown-user",
          match_id: "match-unknown-user",
          account_id: "account-does-not-match",
          score: 12,
          ship_deaths: 3,
          won: true
        },
        headers: internal_headers
    end

    assert_response :not_found
    assert_equal({ "accepted" => false, "error" => "unknown_user" }, JSON.parse(response.body))
  end

  test "POST /internal/player-data/match-results with a missing result_id returns 422" do
    with_internal_token_env do
      post "/internal/player-data/match-results",
        params: {
          match_id: "match-missing-result-id",
          account_id: @user.account_id,
          score: 12,
          ship_deaths: 3,
          won: true
        },
        headers: internal_headers
    end

    assert_response :unprocessable_entity
    assert_equal({ "accepted" => false, "error" => "invalid_input" }, JSON.parse(response.body))
  end

  test "POST /internal/player-data/match-results with a missing match_id returns 422" do
    with_internal_token_env do
      post "/internal/player-data/match-results",
        params: {
          result_id: "result-missing-match-id",
          account_id: @user.account_id,
          score: 12,
          ship_deaths: 3,
          won: true
        },
        headers: internal_headers
    end

    assert_response :unprocessable_entity
    assert_equal({ "accepted" => false, "error" => "invalid_input" }, JSON.parse(response.body))
  end

  test "POST /internal/player-data/match-results with a duplicate result_id returns success without double-counting" do
    with_internal_token_env do
      post "/internal/player-data/match-results",
        params: {
          result_id: "result-duplicate",
          match_id: "match-duplicate",
          account_id: @user.account_id,
          score: 12,
          ship_deaths: 3,
          won: true
        },
        headers: internal_headers
    end

    assert_response :success

    first_body = JSON.parse(response.body)
    assert_equal false, first_body["duplicate"]
    assert_equal 1, first_body["stats"]["games_played"]
    assert_equal 12, first_body["stats"]["total_score"]
    assert_equal 1, first_body["stats"]["wins"]

    with_internal_token_env do
      post "/internal/player-data/match-results",
        params: {
          result_id: "result-duplicate",
          match_id: "match-duplicate",
          account_id: @user.account_id,
          score: 12,
          ship_deaths: 3,
          won: true
        },
        headers: internal_headers
    end

    assert_response :success

    second_body = JSON.parse(response.body)
    assert_equal true, second_body["duplicate"]
    assert_equal 1, second_body["stats"]["games_played"]
    assert_equal 12, second_body["stats"]["total_score"]
    assert_equal 1, second_body["stats"]["wins"]
  end

  private

  def internal_headers(token = "test-internal-token")
    { "Authorization" => "Bearer #{token}" }
  end

  def with_internal_token_env
    previous_value = ENV["GAME_SERVER_INTERNAL_TOKEN"]
    ENV["GAME_SERVER_INTERNAL_TOKEN"] = "test-internal-token"
    yield
  ensure
    if previous_value.nil?
      ENV.delete("GAME_SERVER_INTERNAL_TOKEN")
    else
      ENV["GAME_SERVER_INTERNAL_TOKEN"] = previous_value
    end
  end
end
