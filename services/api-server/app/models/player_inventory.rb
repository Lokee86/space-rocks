class PlayerInventory < ApplicationRecord
  class VersionConflict < StandardError
  end

  belongs_to :user

  validates :user_id, uniqueness: true
  validates :inventory_version, numericality: { only_integer: true, greater_than: 0 }
  validate :inventory_must_be_json_object

  def self.store_for_user!(user:, inventory:, expected_version: nil)
    transaction do
      user.with_lock do
        player_inventory = find_by(user_id: user.id)
        current_version = player_inventory&.inventory_version || 0

        if expected_version && expected_version != current_version
          raise VersionConflict
        end

        next_version = current_version + 1
        versioned_inventory = inventory.deep_stringify_keys.merge("inventory_version" => next_version)

        if player_inventory
          player_inventory.update!(inventory: versioned_inventory, inventory_version: next_version)
        else
          player_inventory = create!(
            user: user,
            inventory: versioned_inventory,
            inventory_version: next_version
          )
        end

        player_inventory
      end
    end
  end

  private

  def inventory_must_be_json_object
    return if inventory.is_a?(Hash)

    errors.add(:inventory, "must be a JSON object")
  end
end
