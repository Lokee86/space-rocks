require "test_helper"

class PasswordCredentialTest < ActiveSupport::TestCase
  test "email is normalized by stripping whitespace and downcasing" do
    user = User.create!(display_name: "One")

    credential = PasswordCredential.new(
      user: user,
      email: "  Example@Email.com  ",
      password: "secret123",
      password_confirmation: "secret123"
    )

    assert credential.valid?
    assert_equal "example@email.com", credential.email
  end

  test "duplicate email is invalid" do
    user_one = User.create!(display_name: "One")
    user_two = User.create!(display_name: "Two")

    PasswordCredential.create!(
      user: user_one,
      email: "duplicate@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )

    duplicate = PasswordCredential.new(
      user: user_two,
      email: "duplicate@example.com",
      password: "anothersecret",
      password_confirmation: "anothersecret"
    )

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:email], "has already been taken"
  end

  test "password authentication succeeds with the correct password" do
    user = User.create!(display_name: "One")

    credential = PasswordCredential.create!(
      user: user,
      email: "auth@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )

    assert credential.authenticate_password("secret123")
  end

  test "password authentication fails with the wrong password" do
    user = User.create!(display_name: "One")

    credential = PasswordCredential.create!(
      user: user,
      email: "auth-fail@example.com",
      password: "secret123",
      password_confirmation: "secret123"
    )

    assert_not credential.authenticate_password("wrong-password")
  end
end
