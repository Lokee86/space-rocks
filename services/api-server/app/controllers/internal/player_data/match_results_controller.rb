module Internal
  module PlayerData
    class MatchResultsController < Internal::BaseController
      def create
        set_api_account_id!(params[:account_id]) if params[:account_id].present?
        unless required_params_present?
          emit_api_event(
            "api_validation_failed",
            context: { "failure_mode" => "missing_match_result_identifiers" },
            specific_failure: true,
            status: :unprocessable_entity
          )
          return render_invalid_input
        end

        emit_match_result_event("match_result_report_started")

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

        unless result.success?
          emit_match_result_event(
            "match_result_report_failed",
            fields: { "failure_stage" => "validation" },
            specific_failure: true,
            status: :unprocessable_entity
          )
          return render_invalid_input
        end

        serialized_stats = PlayerStats::SerializeStats.call(player_stat: result.player_stat)

        if result.duplicate
          emit_match_result_event(
            "match_result_duplicate_suppressed",
            fields: { "duplicate" => true }
          )
        else
          emit_match_result_event(
            "match_result_report_succeeded",
            fields: { "duplicate" => false }
          )
        end

        render json: {
          accepted: true,
          duplicate: result.duplicate,
          stats: serialized_stats
        }
      rescue ActiveRecord::ActiveRecordError
        emit_match_result_event(
          "match_result_report_failed",
          fields: { "failure_stage" => "transaction" },
          specific_failure: true,
          status: 500
        )
        raise
      rescue StandardError
        emit_match_result_event(
          "match_result_report_failed",
          fields: { "failure_stage" => "serialization" },
          specific_failure: true,
          status: 500
        )
        raise
      end

      private

      def emit_match_result_event(event, fields: {}, specific_failure: false, status: nil)
        emit_api_event(
          event,
          context: {
            "account_id" => params[:account_id].to_s.presence,
            "match_id" => params[:match_id].to_s.presence,
            "result_id" => params[:result_id].to_s.presence
          },
          fields: fields,
          specific_failure: specific_failure,
          status: status || (fields["duplicate"].nil? ? nil : :ok)
        )
      end

      def render_unknown_user
        emit_match_result_event(
          "match_result_report_failed",
          fields: { "failure_stage" => "account_lookup" },
          specific_failure: true,
          status: :not_found
        )
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
