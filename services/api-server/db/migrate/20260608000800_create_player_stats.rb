class CreatePlayerStats < ActiveRecord::Migration[8.1]
  def change
    create_table :player_stats do |t|
      t.references :user, null: false, foreign_key: true, index: { unique: true }
      t.integer :total_score, null: false, default: 0
      t.integer :high_score, null: false, default: 0
      t.integer :ship_deaths, null: false, default: 0
      t.integer :games_played, null: false, default: 0
      t.integer :wins, null: false, default: 0

      t.timestamps
    end
  end
end
