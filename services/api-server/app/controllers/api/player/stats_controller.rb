module Api
  module Player
    class StatsController < ApplicationController
      include AuthenticatesBearerToken

      before_action :authenticate_bearer_token!

      def show
        unless current_user
          emit_api_event(
            "api_validation_failed",
            context: { "failure_mode" => "missing_account_identity" },
            specific_failure: true,
            status: :unprocessable_entity
          )
          return render json: { error: "invalid_account" }, status: :unprocessable_entity
        end

        player_stat = current_user.player_stat || current_user.create_player_stat!(zero_stats_attributes)

        render json: {
          stats: PlayerStats::SerializeStats.call(player_stat: player_stat)
        }
      rescue ActiveRecord::ActiveRecordError
        emit_api_event(
          "api_request_failed",
          context: { "failure_mode" => "database_or_stat_creation_failure" },
          specific_failure: true,
          status: 500
        )
        raise
      rescue StandardError
        emit_api_event(
          "api_request_failed",
          context: { "failure_mode" => "stats_serialization_failure" },
          specific_failure: true,
          status: 500
        )
        raise
      end

      private

      def zero_stats_attributes
        {
          total_score: 0,
          high_score: 0,
          ship_deaths: 0,
          games_played: 0,
          wins: 0
        }
      end
    end
  end
end
