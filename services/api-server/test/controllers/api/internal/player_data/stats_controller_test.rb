require "test_helper"

module Api
  module Internal
    module PlayerData
      class StatsControllerTest < ActionDispatch::IntegrationTest
        setup do
          @user = User.create!(display_name: "Ada")
          seed_match_result_for(@user)
        end

        test "POST /api/internal/player-data/stats without an Authorization header returns 401" do
          with_internal_token_env do
            post "/api/internal/player-data/stats", params: {}, as: :json
          end

          assert_response :unauthorized
          assert_openapi_response!
        end

        test "POST /api/internal/player-data/stats with a malformed Authorization header returns 401" do
          with_internal_token_env do
            post "/api/internal/player-data/stats", params: {}, headers: { "Authorization" => "Token test-internal-token" }, as: :json
          end

          assert_response :unauthorized
          assert_openapi_response!
        end

        test "POST /api/internal/player-data/stats with the wrong bearer token returns 401" do
          with_internal_token_env do
            post "/api/internal/player-data/stats", params: {}, headers: internal_headers("wrong-token"), as: :json
          end

          assert_response :unauthorized
          assert_openapi_response!
        end

        test "POST /api/internal/player-data/stats with a valid internal secret and missing account_id returns 422" do
          with_internal_token_env do
            post "/api/internal/player-data/stats", headers: internal_headers, as: :json
          end

          assert_response :unprocessable_entity
          assert_openapi_response!
          assert_equal({ "error" => "invalid_input" }, JSON.parse(response.body))
        end

        test "POST /api/internal/player-data/stats with a valid internal secret and unknown account_id returns 404" do
          with_internal_token_env do
            post "/api/internal/player-data/stats", params: { account_id: "unknown-account" }, headers: internal_headers, as: :json
          end

          assert_response :not_found
          assert_openapi_response!
          assert_equal({ "error" => "unknown_user" }, JSON.parse(response.body))
        end

        test "POST /api/internal/player-data/stats with a valid internal secret and known account_id returns 200" do
          with_internal_token_env do
            post "/api/internal/player-data/stats", params: { account_id: @user.account_id }, headers: internal_headers, as: :json
          end

          assert_response :success
          assert_openapi_contract!

          body = JSON.parse(response.body)

          assert_equal 1, body.keys.size
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

        private

        def seed_match_result_for(user)
          with_internal_token_env do
            post "/internal/player-data/match-results",
              params: {
                result_id: "seed-result-#{user.account_id}",
                match_id: "seed-match-#{user.account_id}",
                account_id: user.account_id,
                score: 12,
                ship_deaths: 3,
                won: true
              },
              headers: internal_headers,
              as: :json
          end

          assert_response :success
          assert_openapi_contract!
        end

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
    end
  end
end
