module Internal
  class BaseController < ApplicationController
    before_action :authenticate_internal_request!

    private

    def authenticate_internal_request!
      expected_token = ENV["GAME_SERVER_INTERNAL_TOKEN"].to_s
      provided_token = bearer_token_from(request.headers["Authorization"])

      return render_unauthorized unless expected_token.present?
      return render_unauthorized unless provided_token
      return render_unauthorized unless expected_token.length == provided_token.length
      return render_unauthorized unless ActiveSupport::SecurityUtils.secure_compare(expected_token, provided_token)
    end

    def bearer_token_from(authorization_header)
      scheme, token = authorization_header.to_s.split(" ", 2)

      return unless scheme == "Bearer" && token.present?

      token
    end

    def render_unauthorized
      render json: { error: "unauthorized" }, status: :unauthorized
    end
  end
end
