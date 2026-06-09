class OauthLoginSession < ApplicationRecord
  belongs_to :user, optional: true

  validates :public_id, presence: true
  validates :provider, presence: true
  validates :poll_secret_digest, presence: true
  validates :status, presence: true
  validates :expires_at, presence: true

  STATUS_PENDING = "pending"
  STATUS_AUTHENTICATED = "authenticated"
  STATUS_CONSUMED = "consumed"
  STATUS_EXPIRED = "expired"

  def expired?
    expires_at.present? && expires_at <= Time.current
  end

  def consumed?
    status == STATUS_CONSUMED || consumed_at.present?
  end

  def pending?
    status == STATUS_PENDING
  end

  def authenticated?
    status == STATUS_AUTHENTICATED
  end

  def usable_for_poll?
    !expired? && !consumed?
  end

  def consume!
    update!(consumed_at: Time.current, status: STATUS_CONSUMED)
  end

  def authenticate!(user)
    update!(user: user, status: STATUS_AUTHENTICATED)
  end
end
