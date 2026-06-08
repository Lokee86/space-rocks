class CreateUserIdentities < ActiveRecord::Migration[8.1]
  def change
    create_table :user_identities do |t|
      t.references :user, null: false, foreign_key: true
      t.string :provider, null: false
      t.string :provider_uid, null: false
      t.string :email

      t.timestamps
    end

    add_index :user_identities, %i[provider provider_uid], unique: true
    add_index :user_identities, :email
  end
end
