class User < ApplicationRecord
  has_one :password_credential, dependent: :destroy
  has_one :player_stat
  has_many :player_match_results, dependent: :destroy
  has_many :user_identities, dependent: :destroy
  has_many :access_tokens, dependent: :destroy

  validates :display_name, presence: true
end
