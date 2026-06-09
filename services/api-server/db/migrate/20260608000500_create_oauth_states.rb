class CreateOauthStates < ActiveRecord::Migration[8.1]
  def change
    create_table :oauth_states do |t|
      t.string :provider, null: false
      t.string :state_digest, null: false
      t.string :redirect_after
      t.datetime :consumed_at
      t.datetime :expires_at, null: false

      t.timestamps
    end

    add_index :oauth_states, :state_digest, unique: true
    add_index :oauth_states, :provider
    add_index :oauth_states, :expires_at
  end
end
