module Auth
  class IssueAccessToken
    def self.call(user:, audience: "api")
      raw_token, access_token = AccessToken.issue_for(user, audience: audience)

      { token: raw_token, access_token: access_token }
    end
  end
end
