require "test_helper"

class Api::Player::StatsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @user = User.create!(display_name: "Ada")
    PasswordCredential.create!(
      user: @user,
      email: "ada@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )
    @raw_token, @access_token = AccessToken.issue_for(@user)
  end

  test "GET /api/player/stats without a token returns 401" do
    get "/api/player/stats"

    assert_response :unauthorized
    assert_openapi_response!
  end

  test "GET /api/player/stats with an invalid token returns 401" do
    get "/api/player/stats", headers: auth_headers("invalid-token")

    assert_response :unauthorized
    assert_openapi_response!
  end

  test "GET /api/player/stats with a valid token returns 200 and creates zero stats" do
    assert_difference -> { PlayerStat.count }, 1 do
      get "/api/player/stats", headers: auth_headers(@raw_token)
    end

    assert_response :success
    assert_openapi_response!

    body = JSON.parse(response.body)
    stats = body["stats"]

    assert_equal 0, stats["total_score"]
    assert_equal 0, stats["high_score"]
    assert_equal 0, stats["ship_deaths"]
    assert_equal 0, stats["games_played"]
    assert_equal 0, stats["wins"]
    assert_nil stats["user_id"]
    assert_nil stats["created_at"]
    assert_nil stats["updated_at"]
    assert_nil stats["token"]
    assert_nil stats["email"]
  end

  test "GET /api/player/stats returns existing stats unchanged" do
    player_stat = PlayerStat.create!(
      user: @user,
      total_score: 12,
      high_score: 9,
      ship_deaths: 3,
      games_played: 4,
      wins: 2
    )

    get "/api/player/stats", headers: auth_headers(@raw_token)

    assert_response :success
    assert_openapi_response!

    body = JSON.parse(response.body)
    stats = body["stats"]

    assert_equal 12, stats["total_score"]
    assert_equal 9, stats["high_score"]
    assert_equal 3, stats["ship_deaths"]
    assert_equal 4, stats["games_played"]
    assert_equal 2, stats["wins"]
    assert_equal player_stat.total_score, stats["total_score"]
    assert_equal player_stat.high_score, stats["high_score"]
    assert_equal player_stat.ship_deaths, stats["ship_deaths"]
    assert_equal player_stat.games_played, stats["games_played"]
    assert_equal player_stat.wins, stats["wins"]
  end

  private

  def auth_headers(raw_token)
    { "Authorization" => "Bearer #{raw_token}" }
  end
end
