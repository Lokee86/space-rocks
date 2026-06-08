module AuthenticatesBearerToken
  extend ActiveSupport::Concern

  included do
    attr_reader :current_user, :current_access_token
  end

  private

  def authenticate_bearer_token!
    access_token = bearer_access_token

    unless access_token
      render json: { error: "invalid_token" }, status: :unauthorized
      return
    end

    @current_access_token = access_token
    @current_user = access_token.user
    @current_access_token.update(last_used_at: Time.current)
  end

  def bearer_access_token
    authorization = request.headers["Authorization"].to_s
    scheme, raw_token = authorization.split(" ", 2)

    return unless scheme == "Bearer" && raw_token.present?

    AccessToken.find_active_by_raw_token(raw_token)
  end
end
