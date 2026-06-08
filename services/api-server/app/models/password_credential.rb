class PasswordCredential < ApplicationRecord
  belongs_to :user

  has_secure_password

  before_validation :normalize_email

  validates :email, presence: true, uniqueness: true
  validates :password_digest, presence: true

  private

  def normalize_email
    self.email = email&.strip&.downcase
  end
end
