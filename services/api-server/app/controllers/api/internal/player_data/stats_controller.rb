module Api
  module Internal
    module PlayerData
      class StatsController < Api::Internal::BaseController
        def create
          return render_invalid_input unless account_id_present?

          user = User.find_by(account_id: params[:account_id])
          return render_unknown_user unless user

          player_stat = user.player_stat || user.create_player_stat!(zero_stats_attributes)

          render json: {
            stats: PlayerStats::SerializeStats.call(player_stat: player_stat)
          }
        end

        private

        def account_id_present?
          params[:account_id].present?
        end

        def render_unknown_user
          render json: { error: "unknown_user" }, status: :not_found
        end

        def render_invalid_input
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
