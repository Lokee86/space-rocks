module Api
  module Auth
    class RegistrationsController < ApplicationController
      include RendersAuthResponse

      def create
        result = ::Auth::RegisterUser.call(
          display_name: registration_params[:display_name],
          email: registration_params[:email],
          password: registration_params[:password]
        )

        if result.success?
          render_auth_success(user: result.user, token: result.token, status: :created)
        else
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
