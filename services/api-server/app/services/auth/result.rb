module Auth
  Result = Struct.new(:user, :token, :error, keyword_init: true) do
    def success?
      error.nil?
    end
  end
end
