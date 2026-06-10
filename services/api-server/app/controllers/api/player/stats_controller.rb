module Api
  module Player
    class StatsController < ApplicationController
      include AuthenticatesBearerToken

      before_action :authenticate_bearer_token!

      def show
        player_stat = current_user.player_stat || current_user.create_player_stat!(zero_stats_attributes)

        render json: {
          stats: PlayerStats::SerializeStats.call(player_stat: player_stat)
        }
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
