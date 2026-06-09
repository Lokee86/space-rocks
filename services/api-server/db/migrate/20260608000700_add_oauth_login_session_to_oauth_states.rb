class AddOauthLoginSessionToOauthStates < ActiveRecord::Migration[8.1]
  def change
    add_reference :oauth_states, :oauth_login_session, null: true, foreign_key: true, index: false
    add_index :oauth_states, :oauth_login_session_id
  end
end
