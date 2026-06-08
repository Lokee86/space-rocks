class UserIdentity < ApplicationRecord
  belongs_to :user

  validates :provider, presence: true
  validates :provider_uid, presence: true, uniqueness: { scope: :provider }
end
