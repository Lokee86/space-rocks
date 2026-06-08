class CreatePasswordCredentials < ActiveRecord::Migration[8.1]
  def change
    create_table :password_credentials do |t|
      t.references :user, null: false, foreign_key: true, index: { unique: true }
      t.string :email, null: false
      t.string :password_digest, null: false

      t.timestamps
    end

    add_index :password_credentials, :email, unique: true
  end
end
