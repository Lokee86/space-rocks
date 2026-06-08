module RendersAuthResponse
  extend ActiveSupport::Concern

  private

  def render_auth_success(user:, token:, status:)
    render json: {
      token: token,
      user: render_auth_user(user)
    }, status: status
  end

  def render_auth_user(user)
    {
      id: user.id,
      display_name: user.display_name,
      email: user.password_credential&.email
    }
  end
end
