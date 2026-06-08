class CreateAccessTokens < ActiveRecord::Migration[8.1]
  def change
    create_table :access_tokens do |t|
      t.references :user, null: false, foreign_key: true
      t.string :token_digest, null: false
      t.string :audience, null: false, default: "api"
      t.datetime :expires_at, null: false
      t.datetime :revoked_at
      t.datetime :last_used_at

      t.timestamps
    end

    add_index :access_tokens, :token_digest, unique: true
    add_index :access_tokens, :audience
    add_index :access_tokens, :expires_at
    add_index :access_tokens, :revoked_at
  end
end
