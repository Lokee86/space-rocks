require "test_helper"
require "uri"

class Api::Auth::DiscordLoginSessionsControllerTest < ActionDispatch::IntegrationTest
  test "POST /api/auth/discord/login_sessions returns the login session payload" do
    client_secret = "super-secret-client-secret"

    with_singleton_method_stub(Auth::Providers::DiscordConfig, :client_id, ->(*args, **kwargs, &block) { "discord-client-id" }) do
      with_singleton_method_stub(Auth::Providers::DiscordConfig, :client_secret, ->(*args, **kwargs, &block) { client_secret }) do
        with_singleton_method_stub(Auth::Providers::DiscordConfig, :redirect_uri, ->(*args, **kwargs, &block) { "https://example.com/api/auth/discord/callback" }) do
          assert_difference "OauthLoginSession.count", 1 do
            assert_difference "OauthState.count", 1 do
              post "/api/auth/discord/login_sessions"
            end
          end
        end
      end
    end

    assert_response :success

    body = JSON.parse(response.body)
    oauth_login_session = OauthLoginSession.order(:id).last
    oauth_state = OauthState.order(:id).last

    assert_equal oauth_login_session.public_id, body["login_session_id"]
    assert_predicate body["poll_secret"], :present?
    assert_predicate body["login_url"], :present?
    assert_predicate body["expires_at"], :present?
    assert_equal oauth_login_session.expires_at.to_i, Time.zone.parse(body["expires_at"]).to_i
    assert_not_equal body["poll_secret"], oauth_login_session.poll_secret_digest
    refute_includes response.body, client_secret

    uri = URI.parse(body["login_url"])
    params = URI.decode_www_form(uri.query).to_h
    assert_equal "discord-client-id", params["client_id"]
    assert_equal "https://example.com/api/auth/discord/callback", params["redirect_uri"]
    assert_equal oauth_state.state_digest, OauthState.digest_for(params["state"])
  end

  test "POST /api/auth/discord/login_sessions/:id/exchange returns pending for an unready session" do
    issued = Auth::OauthLoginSessionIssuer.call(provider: "discord")
    oauth_login_session = issued[:oauth_login_session]

    post "/api/auth/discord/login_sessions/#{oauth_login_session.public_id}/exchange", params: {
      poll_secret: issued[:poll_secret]
    }

    assert_response :accepted
    assert_equal "pending", JSON.parse(response.body)["status"]
  end

  test "POST /api/auth/discord/login_sessions/:id/exchange rejects a wrong poll secret" do
    issued = Auth::OauthLoginSessionIssuer.call(provider: "discord")

    post "/api/auth/discord/login_sessions/#{issued[:oauth_login_session].public_id}/exchange", params: {
      poll_secret: "wrong-secret"
    }

    assert_response :unprocessable_entity
    assert_equal "invalid_login_session", JSON.parse(response.body)["error"]
  end

  test "POST /api/auth/discord/login_sessions/:id/exchange rejects an expired session" do
    issued = Auth::OauthLoginSessionIssuer.call(provider: "discord")
    oauth_login_session = issued[:oauth_login_session]
    oauth_login_session.update!(expires_at: 1.minute.ago)

    post "/api/auth/discord/login_sessions/#{oauth_login_session.public_id}/exchange", params: {
      poll_secret: issued[:poll_secret]
    }

    assert_response :unprocessable_entity
    assert_equal "invalid_login_session", JSON.parse(response.body)["error"]
  end

  test "POST /api/auth/discord/login_sessions/:id/exchange rejects a consumed session" do
    issued = Auth::OauthLoginSessionIssuer.call(provider: "discord")
    oauth_login_session = issued[:oauth_login_session]
    oauth_login_session.consume!

    post "/api/auth/discord/login_sessions/#{oauth_login_session.public_id}/exchange", params: {
      poll_secret: issued[:poll_secret]
    }

    assert_response :unprocessable_entity
    assert_equal "invalid_login_session", JSON.parse(response.body)["error"]
  end

  test "POST /api/auth/discord/login_sessions/:id/exchange issues a bearer token for an authenticated session" do
    user = User.create!(display_name: "Ada Lovelace")
    issued = Auth::OauthLoginSessionIssuer.call(provider: "discord")
    oauth_login_session = issued[:oauth_login_session]
    oauth_login_session.authenticate!(user)

    post "/api/auth/discord/login_sessions/#{oauth_login_session.public_id}/exchange", params: {
      poll_secret: issued[:poll_secret]
    }

    assert_response :success

    body = JSON.parse(response.body)
    assert_predicate body["token"], :present?
    assert_equal user.id, body["user"]["id"]
    assert_equal "Ada Lovelace", body["user"]["display_name"]
    assert_predicate oauth_login_session.reload, :consumed?

    get "/api/auth/me", headers: auth_headers(body["token"])

    assert_response :success
    me_body = JSON.parse(response.body)
    assert_equal user.id, me_body["user"]["id"]
    assert_equal "Ada Lovelace", me_body["user"]["display_name"]
  end

  private

  def auth_headers(raw_token)
    { "Authorization" => "Bearer #{raw_token}" }
  end
end
