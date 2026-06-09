module Auth
  class DiscordController < ApplicationController
    include RendersAuthResponse

    def start
      state_result = Auth::OauthStateIssuer.call(provider: "discord")
      redirect_to Auth::Providers::DiscordAuthorizationUrl.call(
        config: Auth::Providers::DiscordConfig,
        state: state_result[:state]
      ), allow_other_host: true
    end

    def callback
      return render json: { error: :missing_params }, status: :bad_request if params[:code].blank? || params[:state].blank?

      state_result = Auth::OauthStateVerifier.call(provider: "discord", state: params[:state])
      return render json: { error: :invalid_state }, status: :unprocessable_entity unless state_result.success?
      oauth_state = state_result.token

      token_result = Auth::Providers::DiscordTokenExchange.call(code: params[:code])
      return render json: { error: :token_exchange_failed }, status: :bad_gateway unless token_result.success?

      profile_result = Auth::Providers::DiscordCurrentUser.call(access_token: token_result.access_token)
      return render json: { error: :profile_fetch_failed }, status: :bad_gateway unless profile_result.success?

      if oauth_state.oauth_login_session.present?
        user = Auth::OauthResolveUser.call(profile: profile_result.profile)
        return render json: { error: :invalid }, status: :unprocessable_entity unless user

        oauth_state.oauth_login_session.authenticate!(user)
        render json: { message: "You can return to the game." }, status: :ok
      else
        login_result = Auth::OauthLoginUser.call(profile: profile_result.profile)
        if login_result.success?
          render_auth_success(user: login_result.user, token: login_result.token, status: :ok)
        else
          render json: { error: login_result.error }, status: :unprocessable_entity
        end
      end
    end
  end
end
