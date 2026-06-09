class User < ApplicationRecord
  before_validation :assign_account_id, if: -> { account_id.blank? }

  has_one :password_credential, dependent: :destroy
  has_one :player_stat
  has_many :player_match_results, dependent: :destroy
  has_many :user_identities, dependent: :destroy
  has_many :access_tokens, dependent: :destroy

  validates :display_name, presence: true
  validates :account_id, presence: true, uniqueness: true

  private

  def assign_account_id
    self.account_id = SecureRandom.uuid
  end
end
