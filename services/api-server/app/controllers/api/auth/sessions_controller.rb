module Api
  module Auth
    class SessionsController < ApplicationController
      include AuthenticatesBearerToken
      include RendersAuthResponse

      before_action :authenticate_bearer_token!, only: :destroy

      def create
        emit_api_event("auth_flow_started", fields: { "flow" => "login" })
        result = ::Auth::LoginUser.call(
          email: login_params[:email],
          password: login_params[:password]
        )

        if result.success?
          set_api_account_id!(result.user.account_id)
          emit_api_event("auth_succeeded", fields: { "flow" => "login" })
          render_auth_success(user: result.user, token: result.token, status: :ok)
        else
          emit_api_event(
            "auth_failed",
            context: { "reason_code" => api_reason_code(result.error, "authentication_failed") },
            fields: { "flow" => "login" },
            specific_failure: true,
            status: :unauthorized
          )
          render json: { error: result.error }, status: :unauthorized
        end
      end

      def destroy
        current_access_token.update!(revoked_at: Time.current)
        head :no_content
      end

      private

      def login_params
        params.permit(:email, :password)
      end
    end
  end
end
