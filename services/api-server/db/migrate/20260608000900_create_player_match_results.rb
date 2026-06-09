class CreatePlayerMatchResults < ActiveRecord::Migration[8.1]
  def change
    create_table :player_match_results do |t|
      t.string :result_id, null: false
      t.string :match_id, null: false
      t.references :user, null: false, foreign_key: true
      t.integer :score, null: false, default: 0
      t.integer :ship_deaths, null: false, default: 0
      t.boolean :won, null: false, default: false

      t.timestamps
    end

    add_index :player_match_results, :result_id, unique: true
    add_index :player_match_results, :match_id
  end
end
