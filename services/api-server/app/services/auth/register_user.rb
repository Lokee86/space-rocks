module Auth
  class RegisterUser
    def self.call(display_name:, email:, password:)
      user = nil

      ActiveRecord::Base.transaction do
        user = User.create!(display_name: display_name)
        user.create_password_credential!(email: email, password: password, password_confirmation: password)
      end

      token_result = Auth::IssueAccessToken.call(user: user)

      Auth::Result.new(user: user, token: token_result[:token])
    rescue ActiveRecord::RecordInvalid
      Auth::Result.new(error: :invalid)
    end
  end
end
