module Api
  module Auth
    class RegistrationsController < ApplicationController
      include RendersAuthResponse

      def create
        emit_api_event("auth_flow_started", fields: { "flow" => "registration" })
        result = ::Auth::RegisterUser.call(
          display_name: registration_params[:display_name],
          email: registration_params[:email],
          password: registration_params[:password]
        )

        if result.success?
          set_api_account_id!(result.user.account_id)
          emit_api_event("auth_succeeded", fields: { "flow" => "registration" })
          render_auth_success(user: result.user, token: result.token, status: :created)
        else
          emit_api_event(
            "auth_failed",
            context: { "reason_code" => api_reason_code(result.error, "registration_failed") },
            fields: { "flow" => "registration" },
            specific_failure: true,
            status: :unprocessable_entity
          )
          render json: { error: result.error }, status: :unprocessable_entity
        end
      end

      private

      def registration_params
        params.permit(:display_name, :email, :password)
      end
    end
  end
end
