class CreateOauthLoginSessions < ActiveRecord::Migration[8.1]
  def change
    create_table :oauth_login_sessions do |t|
      t.string :public_id, null: false
      t.string :provider, null: false
      t.string :poll_secret_digest, null: false
      t.string :status, null: false
      t.references :user, null: true, foreign_key: true, index: false
      t.datetime :consumed_at
      t.datetime :expires_at, null: false

      t.timestamps
    end

    add_index :oauth_login_sessions, :public_id, unique: true
    add_index :oauth_login_sessions, :poll_secret_digest, unique: true
    add_index :oauth_login_sessions, :user_id
    add_index :oauth_login_sessions, :expires_at
  end
end
