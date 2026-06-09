require "test_helper"
require "net/http"

class Auth::Providers::DiscordCurrentUserTest < ActiveSupport::TestCase
  FakeSuccessResponse = Class.new(Net::HTTPSuccess) do
    attr_reader :body

    def initialize(body)
      @body = body
    end
  end

  test "current-user fetch normalizes id email global_name username and provider" do
    config = Struct.new(:current_user_url).new("https://discord.com/api/users/@me")
    response = FakeSuccessResponse.new(
      {
        id: "12345",
        email: "ada@example.com",
        global_name: "Ada Lovelace",
        username: "ada"
      }.to_json
    )

    with_singleton_method_stub(Auth::Providers::DiscordCurrentUser, :get_current_user, ->(*args, **kwargs, &block) { response }) do
      result = Auth::Providers::DiscordCurrentUser.call(access_token: "discord-access-token", config: config)

      assert_predicate result, :success?
      assert_equal "discord", result.profile.provider
      assert_equal "12345", result.profile.provider_user_id
      assert_equal "ada@example.com", result.profile.email
      assert_equal "Ada Lovelace", result.profile.display_name
      assert_nil result.profile.avatar_url
    end
  end

  test "display_name falls back from global_name to username to Discord User" do
    config = Struct.new(:current_user_url).new("https://discord.com/api/users/@me")

    with_singleton_method_stub(Auth::Providers::DiscordCurrentUser, :get_current_user, ->(*args, **kwargs, &block) { FakeSuccessResponse.new({ id: "1", username: "ada" }.to_json) }) do
      result = Auth::Providers::DiscordCurrentUser.call(access_token: "discord-access-token", config: config)

      assert_equal "ada", result.profile.display_name
    end

    with_singleton_method_stub(Auth::Providers::DiscordCurrentUser, :get_current_user, ->(*args, **kwargs, &block) { FakeSuccessResponse.new({ id: "2" }.to_json) }) do
      result = Auth::Providers::DiscordCurrentUser.call(access_token: "discord-access-token", config: config)

      assert_equal "Discord User", result.profile.display_name
    end
  end
end
