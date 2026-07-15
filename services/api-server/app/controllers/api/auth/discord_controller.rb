module Api
  module Auth
    class DiscordController < ApplicationController
      include RendersAuthResponse

      def start
        emit_api_event("auth_flow_started", fields: { "flow" => "discord_browser" })
        state_result = ::Auth::OauthStateIssuer.call(provider: "discord")
        redirect_to ::Auth::Providers::DiscordAuthorizationUrl.call(
          config: ::Auth::Providers::DiscordConfig,
          state: state_result[:state]
        ), allow_other_host: true
      end

      def callback
        emit_api_event("auth_callback_received", fields: { "provider" => "discord" })
        if params[:code].blank? || params[:state].blank?
          emit_api_event(
            "api_validation_failed",
            context: { "failure_mode" => "missing_oauth_parameters" },
            specific_failure: true,
            status: :bad_request
          )
          return render json: { error: :missing_params }, status: :bad_request
        end

        state_result = ::Auth::OauthStateVerifier.call(provider: "discord", state: params[:state])
        unless state_result.success?
          emit_api_event(
            "auth_failed",
            context: { "reason_code" => "invalid_or_expired_state" },
            fields: { "flow" => "discord_browser" },
            specific_failure: true,
            status: :unprocessable_entity
          )
          return render json: { error: :invalid_state }, status: :unprocessable_entity
        end
        oauth_state = state_result.token

        token_result = ::Auth::Providers::DiscordTokenExchange.call(code: params[:code])
        unless token_result.success?
          emit_api_event(
            "auth_provider_unavailable",
            context: { "failure_mode" => "token_exchange" },
            fields: { "provider" => "discord" },
            specific_failure: true,
            status: :bad_gateway
          )
          return render json: { error: :token_exchange_failed }, status: :bad_gateway
        end

        profile_result = ::Auth::Providers::DiscordCurrentUser.call(access_token: token_result.access_token)
        unless profile_result.success?
          emit_api_event(
            "auth_provider_unavailable",
            context: { "failure_mode" => "profile_fetch" },
            fields: { "provider" => "discord" },
            specific_failure: true,
            status: :bad_gateway
          )
          return render json: { error: :profile_fetch_failed }, status: :bad_gateway
        end

        if oauth_state.oauth_login_session.present?
          user = ::Auth::OauthResolveUser.call(profile: profile_result.profile)
          unless user
            emit_api_event(
              "auth_failed",
              context: { "reason_code" => "user_resolution_failed" },
              fields: { "flow" => "discord_login_session" },
              specific_failure: true,
              status: :unprocessable_entity
            )
            return render json: { error: :invalid }, status: :unprocessable_entity
          end

          oauth_state.oauth_login_session.authenticate!(user)
          set_api_account_id!(user.account_id)
          emit_api_event("auth_succeeded", fields: { "flow" => "discord_login_session" })
          render json: { message: "You can return to the game." }, status: :ok
        else
          login_result = ::Auth::OauthLoginUser.call(profile: profile_result.profile)
          if login_result.success?
            set_api_account_id!(login_result.user.account_id)
            emit_api_event("auth_succeeded", fields: { "flow" => "discord_browser" })
            render_auth_success(user: login_result.user, token: login_result.token, status: :ok)
          else
            emit_api_event(
              "auth_failed",
              context: { "reason_code" => api_reason_code(login_result.error, "authentication_failed") },
              fields: { "flow" => "discord_browser" },
              specific_failure: true,
              status: :unprocessable_entity
            )
            render json: { error: login_result.error }, status: :unprocessable_entity
          end
        end
      end
    end
  end
end
