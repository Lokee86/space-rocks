module Auth
  module Providers
    class DiscordConfig
      AUTHORIZATION_URL = "https://discord.com/oauth2/authorize".freeze
      TOKEN_URL = "https://discord.com/api/oauth2/token".freeze
      CURRENT_USER_URL = "https://discord.com/api/users/@me".freeze

      def self.authorization_url
        AUTHORIZATION_URL
      end

      def self.token_url
        TOKEN_URL
      end

      def self.current_user_url
        CURRENT_USER_URL
      end

      def self.client_id
        ENV.fetch("DISCORD_CLIENT_ID")
      end

      def self.client_secret
        ENV.fetch("DISCORD_CLIENT_SECRET")
      end

      def self.redirect_uri
        ENV.fetch("DISCORD_REDIRECT_URI")
      end
    end
  end
end
