module Auth
  module Providers
    class ProviderProfile
      attr_reader :provider, :provider_user_id, :email, :display_name, :avatar_url

      def initialize(provider:, provider_user_id:, email:, display_name:, avatar_url:)
        @provider = provider
        @provider_user_id = provider_user_id
        @email = email
        @display_name = display_name
        @avatar_url = avatar_url

        freeze
      end
    end
  end
end
