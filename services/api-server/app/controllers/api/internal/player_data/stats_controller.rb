module Api
  module Internal
    module PlayerData
      class StatsController < Api::Internal::BaseController
        def create
          set_api_account_id!(params[:account_id]) if params[:account_id].present?
          return render_invalid_input unless account_id_present?

          user = User.find_by(account_id: params[:account_id])
          return render_unknown_user unless user

          player_stat = user.player_stat || user.create_player_stat!(zero_stats_attributes)

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

        def account_id_present?
          params[:account_id].present?
        end

        def render_unknown_user
          emit_api_event(
            "api_request_failed",
            context: { "failure_mode" => "unknown_user" },
            specific_failure: true,
            status: :not_found
          )
          render json: { error: "unknown_user" }, status: :not_found
        end

        def render_invalid_input
          emit_api_event(
            "api_validation_failed",
            context: { "failure_mode" => "missing_account_identity" },
            specific_failure: true,
            status: :unprocessable_entity
          )
          render json: { error: "invalid_input" }, status: :unprocessable_entity
        end

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
end
