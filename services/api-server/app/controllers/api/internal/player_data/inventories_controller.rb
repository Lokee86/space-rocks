module Api
  module Internal
    module PlayerData
      class InventoriesController < Api::Internal::BaseController
        def load
          set_api_account_id!(params[:account_id]) if params[:account_id].present?
          return render_invalid_input unless account_id_present?

          user = User.find_by(account_id: params[:account_id])
          return render_unknown_user unless user

          player_inventory = user.player_inventory
          return render json: { found: false } unless player_inventory

          render json: {
            found: true,
            inventory: player_inventory.inventory,
            inventory_version: player_inventory.inventory_version
          }
        end

        def store
          set_api_account_id!(params[:account_id]) if params[:account_id].present?
          return render_invalid_input unless account_id_present? && inventory_object?

          expected_version = normalized_expected_version
          return render_invalid_input if expected_version == :invalid

          user = User.find_by(account_id: params[:account_id])
          return render_unknown_user unless user

          player_inventory = PlayerInventory.store_for_user!(
            user: user,
            inventory: normalized_inventory,
            expected_version: expected_version
          )

          render json: {
            inventory: player_inventory.inventory,
            inventory_version: player_inventory.inventory_version
          }
        rescue PlayerInventory::VersionConflict
          render json: { error: "inventory_version_conflict" }, status: :conflict
        end

        private

        def account_id_present?
          params[:account_id].present?
        end

        def inventory_object?
          params[:inventory].respond_to?(:to_unsafe_h) || params[:inventory].is_a?(Hash)
        end

        def normalized_inventory
          inventory = params[:inventory]
          return inventory.to_unsafe_h if inventory.respond_to?(:to_unsafe_h)

          inventory
        end

        def normalized_expected_version
          return nil unless params.key?(:expected_version) && params[:expected_version].present?

          version = Integer(params[:expected_version].to_s, 10)
          return :invalid if version.negative?

          version
        rescue ArgumentError, TypeError
          :invalid
        end

        def render_unknown_user
          render json: { error: "unknown_user" }, status: :not_found
        end

        def render_invalid_input
          render json: { error: "invalid_input" }, status: :unprocessable_entity
        end
      end
    end
  end
end
