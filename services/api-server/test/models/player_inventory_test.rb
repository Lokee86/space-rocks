require "test_helper"

class PlayerInventoryTest < ActiveSupport::TestCase
  test "stores a JSON object for a user" do
    user = User.create!(display_name: "Ada")
    inventory = { "schema_version" => 1, "owned_ships" => [] }

    player_inventory = PlayerInventory.create!(user: user, inventory: inventory)

    assert_equal player_inventory, user.reload.player_inventory
    assert_equal inventory, player_inventory.inventory
    assert_equal 1, player_inventory.inventory_version
  end

  test "a user can have only one inventory" do
    user = User.create!(display_name: "Ada")
    PlayerInventory.create!(user: user, inventory: {})

    duplicate = PlayerInventory.new(user: user, inventory: {})

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:user_id], "has already been taken"
  end

  test "inventory must be a JSON object" do
    player_inventory = PlayerInventory.new(
      user: User.create!(display_name: "Ada"),
      inventory: []
    )

    assert_not player_inventory.valid?
    assert_includes player_inventory.errors[:inventory], "must be a JSON object"
  end

  test "store increments the version and enforces the expected version" do
    user = User.create!(display_name: "Ada")

    stored = PlayerInventory.store_for_user!(user: user, inventory: { "revision" => 1 }, expected_version: 0)
    updated = PlayerInventory.store_for_user!(user: user, inventory: { "revision" => 2 }, expected_version: 1)

    assert_equal 1, stored.inventory_version
    assert_equal 2, updated.inventory_version
    assert_equal({ "revision" => 2, "inventory_version" => 2 }, updated.reload.inventory)
    assert_raises(PlayerInventory::VersionConflict) do
      PlayerInventory.store_for_user!(user: user, inventory: { "revision" => 3 }, expected_version: 1)
    end
  end
end
