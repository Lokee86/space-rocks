module Internal
  class BaseController < ApplicationController
    before_action :authenticate_internal_request!

    private

    def authenticate_internal_request!
      expected_token = ENV["GAME_SERVER_INTERNAL_TOKEN"].to_s
      authorization_header = request.headers["Authorization"].to_s

      return render_unauthorized(reason_code: "internal_token_not_configured") unless expected_token.present?
      return render_unauthorized(reason_code: "authorization_missing") if authorization_header.blank?

      scheme, token = authorization_header.to_s.split(" ", 2)
      return render_unauthorized(reason_code: "authorization_scheme_invalid") unless scheme == "Bearer" && token.present?
      return render_unauthorized(reason_code: "internal_token_invalid") unless expected_token.length == token.length
      return render_unauthorized(reason_code: "internal_token_invalid") unless ActiveSupport::SecurityUtils.secure_compare(expected_token, token)
    end

    def render_unauthorized(reason_code:)
      emit_api_event(
        "api_unauthorized",
        context: { "reason_code" => reason_code },
        specific_failure: true,
        status: :unauthorized
      )
      render json: { error: "unauthorized" }, status: :unauthorized
    end
  end
end
