module Internal
  module Auth
    class VerifyTokensController < Internal::BaseController
      def create
        result = ::Auth::VerifyAccessToken.call(raw_token: params[:token])

        if result.success?
          render json: {
            valid: true,
            user: {
              id: result.user.id,
              display_name: result.user.display_name
            }
          }, status: :ok
        else
          render json: { valid: false }, status: :ok
        end
      end
    end
  end
end
