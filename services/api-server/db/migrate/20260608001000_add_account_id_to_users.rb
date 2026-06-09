require "securerandom"

class AddAccountIdToUsers < ActiveRecord::Migration[8.1]
  def up
    add_column :users, :account_id, :string

    User.reset_column_information
    User.find_each do |user|
      user.update_columns(account_id: SecureRandom.uuid)
    end

    change_column_null :users, :account_id, false
    add_index :users, :account_id, unique: true
  end

  def down
    remove_index :users, :account_id
    remove_column :users, :account_id
  end
end
