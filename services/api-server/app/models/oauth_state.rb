class OauthState < ApplicationRecord
  validates :provider, presence: true
  validates :state_digest, presence: true
  validates :expires_at, presence: true

  def self.digest_for(raw_state)
    Digest::SHA256.hexdigest(raw_state.to_s)
  end

  def expired?
    expires_at.present? && expires_at <= Time.current
  end

  def consumed?
    consumed_at.present?
  end

  def usable?
    !expired? && !consumed?
  end

  def consume!
    update!(consumed_at: Time.current)
  end
end
