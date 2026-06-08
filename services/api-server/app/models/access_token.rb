class AccessToken < ApplicationRecord
  TOKEN_BYTES = 32

  belongs_to :user

  validates :token_digest, presence: true, uniqueness: true
  validates :audience, presence: true
  validates :expires_at, presence: true

  def self.digest_for(raw_token)
    Digest::SHA256.hexdigest(raw_token.to_s)
  end

  def self.issue_for(user, audience: "api", expires_at: 30.days.from_now)
    raw_token = SecureRandom.hex(TOKEN_BYTES)
    record = create!(
      user: user,
      token_digest: digest_for(raw_token),
      audience: audience,
      expires_at: expires_at
    )

    [raw_token, record]
  end

  def self.find_active_by_raw_token(raw_token)
    token = find_by(token_digest: digest_for(raw_token))
    return unless token&.active?

    token
  end

  def revoked?
    revoked_at.present?
  end

  def expired?
    expires_at.present? && expires_at <= Time.current
  end

  def active?
    !revoked? && !expired?
  end
end
