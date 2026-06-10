require "test_helper"
require "uri"

class Api::Auth::DiscordControllerTest < ActionDispatch::IntegrationTest
  test "GET /api/auth/discord/start redirects to Discord and creates an oauth state" do
    with_singleton_method_stub(Auth::Providers::DiscordConfig, :client_id, ->(*args, **kwargs, &block) { "discord-client-id" }) do
      with_singleton_method_stub(Auth::Providers::DiscordConfig, :redirect_uri, ->(*args, **kwargs, &block) { "https://example.com/api/auth/discord/callback" }) do
        assert_difference "OauthState.count", 1 do
          get "/api/auth/discord/start"
        end
      end
    end

    assert_response :redirect

    uri = URI.parse(response.redirect_url)
    params = URI.decode_www_form(uri.query).to_h

    assert_equal Auth::Providers::DiscordConfig.authorization_url, "#{uri.scheme}://#{uri.host}#{uri.path}"
    assert_equal "discord-client-id", params["client_id"]
    assert_equal "https://example.com/api/auth/discord/callback", params["redirect_uri"]
    assert_equal "code", params["response_type"]
    assert_equal "identify email", params["scope"]
    assert_predicate params["state"], :present?
  end

  test "GET /api/auth/discord/callback returns a JSON error when code is missing" do
    state = Auth::OauthStateIssuer.call(provider: "discord")[:state]

    get "/api/auth/discord/callback", params: { state: state }

    assert_response :bad_request
    assert_equal "missing_params", JSON.parse(response.body)["error"]
  end

  test "GET /api/auth/discord/callback returns a JSON error when state is missing" do
    get "/api/auth/discord/callback", params: { code: "auth-code" }

    assert_response :bad_request
    assert_equal "missing_params", JSON.parse(response.body)["error"]
  end

  test "GET /api/auth/discord/callback returns a JSON error for invalid state" do
    get "/api/auth/discord/callback", params: { code: "auth-code", state: "bogus-state" }

    assert_response :unprocessable_entity
    assert_equal "invalid_state", JSON.parse(response.body)["error"]
  end

  test "GET /api/auth/discord/callback creates a user and returns the normal auth response" do
    profile_result = Struct.new(:success?, :profile).new(
      true,
      Auth::Providers::ProviderProfile.new(
        provider: "discord",
        provider_user_id: "discord-user-1",
        email: nil,
        display_name: "Ada Lovelace",
        avatar_url: nil
      )
    )
    token_result = Struct.new(:success?, :access_token).new(true, "discord-access-token")
    state = Auth::OauthStateIssuer.call(provider: "discord")[:state]

    with_singleton_method_stub(Auth::Providers::DiscordTokenExchange, :call, ->(*args, **kwargs, &block) { token_result }) do
      with_singleton_method_stub(Auth::Providers::DiscordCurrentUser, :call, ->(*args, **kwargs, &block) { profile_result }) do
        assert_difference "User.count", 1 do
          assert_difference "UserIdentity.count", 1 do
            get "/api/auth/discord/callback", params: { code: "auth-code", state: state }
          end
        end
      end
    end

    assert_response :ok

    body = JSON.parse(response.body)

    assert_predicate body["token"], :present?
    assert_equal "Ada Lovelace", body["user"]["display_name"]
    assert_nil body["user"]["email"]

    user = User.find(body["user"]["id"])

    get "/api/auth/me", headers: auth_headers(body["token"])

    assert_response :ok
    me_body = JSON.parse(response.body)
    assert_equal user.id, me_body["user"]["id"]
    assert_equal "Ada Lovelace", me_body["user"]["display_name"]
    assert_nil me_body["user"]["email"]
  end

  test "GET /api/auth/discord/callback reuses the same user on repeated logins" do
    profile_result = Struct.new(:success?, :profile).new(
      true,
      Auth::Providers::ProviderProfile.new(
        provider: "discord",
        provider_user_id: "discord-user-1",
        email: nil,
        display_name: "Ada Lovelace",
        avatar_url: nil
      )
    )
    token_result = Struct.new(:success?, :access_token).new(true, "discord-access-token")

    with_singleton_method_stub(Auth::Providers::DiscordTokenExchange, :call, ->(*args, **kwargs, &block) { token_result }) do
      with_singleton_method_stub(Auth::Providers::DiscordCurrentUser, :call, ->(*args, **kwargs, &block) { profile_result }) do
        first_state = Auth::OauthStateIssuer.call(provider: "discord")[:state]
        second_state = Auth::OauthStateIssuer.call(provider: "discord")[:state]

        get "/api/auth/discord/callback", params: { code: "auth-code", state: first_state }
        first_user_id = JSON.parse(response.body)["user"]["id"]

        get "/api/auth/discord/callback", params: { code: "auth-code", state: second_state }
        second_user_id = JSON.parse(response.body)["user"]["id"]

        assert_equal first_user_id, second_user_id
        assert_equal 1, User.where(display_name: "Ada Lovelace").count
        assert_equal 1, UserIdentity.where(provider: "discord", provider_uid: "discord-user-1").count
      end
    end
  end

  private

  def auth_headers(raw_token)
    { "Authorization" => "Bearer #{raw_token}" }
  end
end
