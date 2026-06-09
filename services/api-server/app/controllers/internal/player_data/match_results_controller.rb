module Internal
  module PlayerData
    class MatchResultsController < Internal::BaseController
      def create
        return render_invalid_input unless required_params_present?

        user = User.find_by(account_id: params[:account_id])
        return render_unknown_user unless user

        result = PlayerStats::ApplyMatchResult.call(
          user: user,
          result_id: params[:result_id],
          match_id: params[:match_id],
          score: normalized_score,
          ship_deaths: normalized_ship_deaths,
          won: normalized_won
        )

        return render_invalid_input unless result.success?

        render json: {
          accepted: true,
          duplicate: result.duplicate,
          stats: PlayerStats::SerializeStats.call(player_stat: result.player_stat)
        }
      end

      private

      def render_unknown_user
        render json: { accepted: false, error: "unknown_user" }, status: :not_found
      end

      def render_invalid_input
        render json: { accepted: false, error: "invalid_input" }, status: :unprocessable_entity
      end

      def required_params_present?
        params[:result_id].present? && params[:match_id].present? && params[:account_id].present?
      end

      def normalized_score
        params[:score].to_i
      end

      def normalized_ship_deaths
        params[:ship_deaths].to_i
      end

      def normalized_won
        value = params[:won]
        return true if value == true || value == "true" || value == "1" || value == 1

        false
      end
    end
  end
end
