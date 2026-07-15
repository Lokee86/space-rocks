module Internal
  module Auth
    class VerifyTokensController < Internal::BaseController
      def create
        result = ::Auth::VerifyAccessToken.call(raw_token: params[:token])

        if result.success?
          set_api_account_id!(result.user.account_id)
          emit_api_event("auth_succeeded", fields: { "flow" => "token_verification" })
          render json: {
            valid: true,
            user: {
              id: result.user.id,
              account_id: result.user.account_id,
              display_name: result.user.display_name
            }
          }, status: :ok
        else
          emit_api_event(
            "auth_failed",
            context: { "reason_code" => "invalid_access_token" },
            fields: { "flow" => "token_verification" },
            specific_failure: true,
            status: :ok
          )
          render json: { valid: false }, status: :ok
        end
      end
    end
  end
end
