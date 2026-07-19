class CreatePlayerInventories < ActiveRecord::Migration[8.1]
  def change
    create_table :player_inventories do |t|
      t.references :user, null: false, foreign_key: true, index: { unique: true }
      t.jsonb :inventory, null: false
      t.integer :inventory_version, null: false, default: 1

      t.timestamps
    end
  end
end
