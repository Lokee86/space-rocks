require "json"
require "net/http"
require "uri"

module Auth
  module Providers
    class DiscordTokenExchange
      Result = Struct.new(:access_token, :error, keyword_init: true) do
        def success?
          error.nil?
        end
      end

      HTTP_ERROR = :http_error
      JSON_ERROR = :invalid_json
      INVALID_RESPONSE_ERROR = :invalid_response

      def self.call(code:, config: DiscordConfig)
        response = post_token_request(code: code, config: config)
        return Result.new(error: HTTP_ERROR) unless response.is_a?(Net::HTTPSuccess)

        payload = JSON.parse(response.body)
        access_token = payload["access_token"]
        return Result.new(error: INVALID_RESPONSE_ERROR) if access_token.to_s.empty?

        Result.new(access_token: access_token)
      rescue JSON::ParserError
        Result.new(error: JSON_ERROR)
      rescue StandardError
        Result.new(error: HTTP_ERROR)
      end

      def self.post_token_request(code:, config:)
        uri = URI.parse(config.token_url)
        request = Net::HTTP::Post.new(uri)
        request.set_form_data(
          client_id: config.client_id,
          client_secret: config.client_secret,
          grant_type: "authorization_code",
          code: code,
          redirect_uri: config.redirect_uri
        )

        Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == "https") do |http|
          http.request(request)
        end
      end
      private_class_method :post_token_request
    end
  end
end
