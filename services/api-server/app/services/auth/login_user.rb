module Auth
  class LoginUser
    def self.call(email:, password:)
      normalized_email = email.to_s.strip.downcase
      password_credential = PasswordCredential.find_by(email: normalized_email)

      return Auth::Result.new(error: :invalid_credentials) unless password_credential&.authenticate_password(password)

      token_result = Auth::IssueAccessToken.call(user: password_credential.user)

      Auth::Result.new(user: password_credential.user, token: token_result[:token])
    end
  end
end
