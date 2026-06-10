module Api
  module Auth
    class SessionsController < ApplicationController
      include AuthenticatesBearerToken
      include RendersAuthResponse

      before_action :authenticate_bearer_token!, only: :destroy

      def create
        result = Auth::LoginUser.call(
          email: login_params[:email],
          password: login_params[:password]
        )

        if result.success?
          render_auth_success(user: result.user, token: result.token, status: :ok)
        else
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
