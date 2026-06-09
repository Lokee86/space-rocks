module Auth
  class OauthLoginSessionIssuer
    PUBLIC_ID_BYTES = 16
    POLL_SECRET_BYTES = 32

    def self.call(provider: "discord")
      public_id = SecureRandom.hex(PUBLIC_ID_BYTES)
      raw_poll_secret = SecureRandom.hex(POLL_SECRET_BYTES)

      oauth_login_session = OauthLoginSession.create!(
        public_id: public_id,
        provider: provider,
        poll_secret_digest: Digest::SHA256.hexdigest(raw_poll_secret),
        status: OauthLoginSession::STATUS_PENDING,
        expires_at: 10.minutes.from_now
      )

      { poll_secret: raw_poll_secret, oauth_login_session: oauth_login_session }
    end
  end
end
