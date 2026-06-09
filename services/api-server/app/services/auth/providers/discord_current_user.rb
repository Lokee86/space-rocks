require "json"
require "net/http"
require "uri"

module Auth
  module Providers
    class DiscordCurrentUser
      Result = Struct.new(:profile, :error, keyword_init: true) do
        def success?
          error.nil?
        end
      end

      HTTP_ERROR = :http_error
      JSON_ERROR = :invalid_json
      INVALID_RESPONSE_ERROR = :invalid_response

      def self.call(access_token:, config: DiscordConfig)
        response = get_current_user(access_token: access_token, config: config)
        return Result.new(error: HTTP_ERROR) unless response.is_a?(Net::HTTPSuccess)

        payload = JSON.parse(response.body)
        provider_user_id = payload["id"]
        return Result.new(error: INVALID_RESPONSE_ERROR) if provider_user_id.to_s.empty?

        Result.new(
          profile: ProviderProfile.new(
            provider: "discord",
            provider_user_id: provider_user_id,
            email: payload["email"],
            display_name: payload["global_name"].presence || payload["username"].presence || "Discord User",
            avatar_url: nil
          )
        )
      rescue JSON::ParserError
        Result.new(error: JSON_ERROR)
      rescue StandardError
        Result.new(error: HTTP_ERROR)
      end

      def self.get_current_user(access_token:, config:)
        uri = URI.parse(config.current_user_url)
        request = Net::HTTP::Get.new(uri)
        request["Authorization"] = "Bearer #{access_token}"

        Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == "https") do |http|
          http.request(request)
        end
      end
      private_class_method :get_current_user
    end
  end
end
