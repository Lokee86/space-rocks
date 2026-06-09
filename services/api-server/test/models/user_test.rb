require "test_helper"

class UserTest < ActiveSupport::TestCase
  test "new user gets an account_id when saved" do
    user = User.create!(display_name: "One")

    assert_predicate user.account_id, :present?
  end

  test "account_id is present" do
    user = User.create!(display_name: "Two")

    assert_predicate user.account_id, :present?
  end

  test "account_id is unique" do
    user = User.create!(display_name: "One")
    duplicate = User.new(display_name: "Two", account_id: user.account_id)

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:account_id], "has already been taken"
  end

  test "account_id is a string" do
    user = User.create!(display_name: "Three")

    assert_kind_of String, user.account_id
  end
end
