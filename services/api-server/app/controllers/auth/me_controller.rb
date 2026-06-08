module Auth
  class MeController < ApplicationController
    include AuthenticatesBearerToken

    before_action :authenticate_bearer_token!

    def show
      render json: {
        user: {
          id: current_user.id,
          display_name: current_user.display_name,
          email: current_user.password_credential&.email
        }
      }
    end
  end
end
