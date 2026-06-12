module Api
  module Auth
    class DiscordLoginSessionsController < ApplicationController
      include RendersAuthResponse

      def create
        login_session_result = ::Auth::OauthLoginSessionIssuer.call(provider: "discord")
        state_result = ::Auth::OauthStateIssuer.call(
          provider: "discord",
          oauth_login_session: login_session_result[:oauth_login_session]
        )
        login_url = ::Auth::Providers::DiscordAuthorizationUrl.call(
          config: ::Auth::Providers::DiscordConfig,
          state: state_result[:state]
        )

        render json: {
          login_session_id: login_session_result[:oauth_login_session].public_id,
          poll_secret: login_session_result[:poll_secret],
          login_url: login_url,
          expires_at: login_session_result[:oauth_login_session].expires_at
        }
      end

      def exchange
        return render json: { error: :missing_params }, status: :bad_request if params[:poll_secret].blank?

        oauth_login_session = OauthLoginSession.find_by(public_id: params[:id])
        return render_invalid_exchange unless oauth_login_session
        return render_invalid_exchange if oauth_login_session.expired? || oauth_login_session.consumed?
        return render_invalid_exchange unless valid_poll_secret?(oauth_login_session)
        return render_pending_exchange if oauth_login_session.pending?
        return render_invalid_exchange unless oauth_login_session.authenticated? && oauth_login_session.user.present?

        token_result = nil
        ActiveRecord::Base.transaction do
          token_result = ::Auth::IssueAccessToken.call(user: oauth_login_session.user)
          oauth_login_session.consume!
        end

        render_auth_success(user: oauth_login_session.user, token: token_result[:token], status: :ok)
      end

      private

      def valid_poll_secret?(oauth_login_session)
        expected_digest = Digest::SHA256.hexdigest(params[:poll_secret].to_s)
        ActiveSupport::SecurityUtils.secure_compare(
          expected_digest,
          oauth_login_session.poll_secret_digest
        )
      end

      def render_pending_exchange
        render json: { status: "pending" }, status: :accepted
      end

      def render_invalid_exchange
        render json: { error: :invalid_login_session }, status: :unprocessable_entity
      end
    end
  end
end
