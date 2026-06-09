module AuthenticatesBearerToken
  extend ActiveSupport::Concern

  included do
    attr_reader :current_user, :current_access_token
  end

  private

  def authenticate_bearer_token!
    raw_token = bearer_raw_token
    result = ::Auth::VerifyAccessToken.call(raw_token: raw_token)

    unless result.success?
      render json: { error: "invalid_token" }, status: :unauthorized
      return
    end

    @current_access_token = result.token
    @current_user = result.user
  end

  def bearer_raw_token
    authorization = request.headers["Authorization"].to_s
    scheme, raw_token = authorization.split(" ", 2)

    return unless scheme == "Bearer" && raw_token.present?

    raw_token
  end
end
