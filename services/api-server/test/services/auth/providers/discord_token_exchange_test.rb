require "test_helper"
require "net/http"

class Auth::Providers::DiscordTokenExchangeTest < ActiveSupport::TestCase
  FakeSuccessResponse = Class.new(Net::HTTPSuccess) do
    attr_reader :body

    def initialize(body)
      @body = body
    end
  end

  test "token exchange parses a successful mocked Discord token response" do
    config = Struct.new(:token_url, :client_id, :client_secret, :redirect_uri).new(
      "https://discord.com/api/oauth2/token",
      "client-id",
      "client-secret",
      "https://example.com/api/auth/discord/callback"
    )

    response = FakeSuccessResponse.new({ access_token: "discord-access-token" }.to_json)

    with_singleton_method_stub(Auth::Providers::DiscordTokenExchange, :post_token_request, ->(*args, **kwargs, &block) { response }) do
      result = Auth::Providers::DiscordTokenExchange.call(code: "auth-code", config: config)

      assert_predicate result, :success?
      assert_equal "discord-access-token", result.access_token
    end
  end

  test "token exchange returns failure on mocked HTTP failure" do
    config = Struct.new(:token_url, :client_id, :client_secret, :redirect_uri).new(
      "https://discord.com/api/oauth2/token",
      "client-id",
      "client-secret",
      "https://example.com/api/auth/discord/callback"
    )

    with_singleton_method_stub(Auth::Providers::DiscordTokenExchange, :post_token_request, ->(*args, **kwargs, &block) { Object.new }) do
      result = Auth::Providers::DiscordTokenExchange.call(code: "auth-code", config: config)

      assert_not result.success?
      assert_equal Auth::Providers::DiscordTokenExchange::HTTP_ERROR, result.error
    end
  end
end
