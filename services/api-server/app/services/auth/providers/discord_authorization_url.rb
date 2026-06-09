require "uri"

module Auth
  module Providers
    class DiscordAuthorizationUrl
      def self.call(config:, state:)
        uri = URI.parse(config.authorization_url)
        uri.query = URI.encode_www_form(
          client_id: config.client_id,
          redirect_uri: config.redirect_uri,
          response_type: "code",
          scope: "identify email",
          state: state
        )
        uri.to_s
      end
    end
  end
end
