require "test_helper"

class Auth::OauthLoginUserTest < ActiveSupport::TestCase
  test "first Discord login creates a user, user_identity, and access token" do
    profile = Auth::Providers::ProviderProfile.new(
      provider: "discord",
      provider_user_id: "discord-user-1",
      email: "ada@example.com",
      display_name: "Ada Lovelace",
      avatar_url: nil
    )

    assert_difference "User.count", 1 do
      assert_difference "UserIdentity.count", 1 do
        assert_difference "AccessToken.count", 1 do
          result = Auth::OauthLoginUser.call(profile: profile)

          assert_predicate result, :success?
          assert_equal "Ada Lovelace", result.user.display_name
          assert_equal profile.provider, result.user.user_identities.first.provider
          assert_equal profile.provider_user_id, result.user.user_identities.first.provider_uid
          assert_equal "ada@example.com", result.user.user_identities.first.email
          assert_predicate result.token, :present?
        end
      end
    end
  end

  test "second login with the same provider and provider_user_id reuses the same user" do
    profile = Auth::Providers::ProviderProfile.new(
      provider: "discord",
      provider_user_id: "discord-user-1",
      email: nil,
      display_name: "Ada Lovelace",
      avatar_url: nil
    )

    first_result = Auth::OauthLoginUser.call(profile: profile)
    second_result = Auth::OauthLoginUser.call(profile: profile)

    assert_equal first_result.user.id, second_result.user.id
    assert_equal 1, UserIdentity.where(provider: "discord", provider_uid: "discord-user-1").count
  end

  test "email may be nil" do
    profile = Auth::Providers::ProviderProfile.new(
      provider: "discord",
      provider_user_id: "discord-user-2",
      email: nil,
      display_name: "Discord User",
      avatar_url: nil
    )

    result = Auth::OauthLoginUser.call(profile: profile)

    assert_predicate result, :success?
    assert_nil result.user.user_identities.first.email
  end
end
