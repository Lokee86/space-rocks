require "test_helper"

module Api
  module Internal
    module PlayerData
      class InventoriesControllerTest < ActionDispatch::IntegrationTest
        setup do
          @user = User.create!(display_name: "Ada")
        end

        test "POST /api/internal/player-data/inventory/load requires the internal bearer token" do
          post "/api/internal/player-data/inventory/load", params: { account_id: @user.account_id }, as: :json

          assert_response :unauthorized
        end

        test "load returns found false without creating starter inventory" do
          with_internal_token_env do
            post "/api/internal/player-data/inventory/load",
              params: { account_id: @user.account_id },
              headers: internal_headers,
              as: :json
          end

          assert_response :success
          assert_equal({ "found" => false }, JSON.parse(response.body))
          assert_nil @user.reload.player_inventory
        end

        test "store creates and returns the supplied inventory and version" do
          inventory = { schema_version: 1, owned_ships: [] }

          with_internal_token_env do
            post "/api/internal/player-data/inventory/store",
              params: { account_id: @user.account_id, inventory: inventory, expected_version: 0 },
              headers: internal_headers,
              as: :json
          end

          assert_response :success
          assert_equal(
            {
              "inventory" => { "schema_version" => 1, "owned_ships" => [], "inventory_version" => 1 },
              "inventory_version" => 1
            },
            JSON.parse(response.body)
          )
        end

        test "store replaces the inventory and increments the version" do
          PlayerInventory.create!(user: @user, inventory: { "revision" => 1 })

          with_internal_token_env do
            post "/api/internal/player-data/inventory/store",
              params: {
                account_id: @user.account_id,
                inventory: { revision: 2 },
                expected_version: 1
              },
              headers: internal_headers,
              as: :json
          end

          assert_response :success
          body = JSON.parse(response.body)
          assert_equal({ "revision" => 2, "inventory_version" => 2 }, body["inventory"])
          assert_equal 2, body["inventory_version"]
        end

        test "store returns a stable conflict when the expected version is stale" do
          PlayerInventory.create!(user: @user, inventory: { "revision" => 1 })

          with_internal_token_env do
            post "/api/internal/player-data/inventory/store",
              params: {
                account_id: @user.account_id,
                inventory: { revision: 2 },
                expected_version: 0
              },
              headers: internal_headers,
              as: :json
          end

          assert_response :conflict
          assert_equal({ "error" => "inventory_version_conflict" }, JSON.parse(response.body))
          assert_equal({ "revision" => 1 }, @user.reload.player_inventory.inventory)
        end

        test "store rejects a non-object inventory" do
          with_internal_token_env do
            post "/api/internal/player-data/inventory/store",
              params: { account_id: @user.account_id, inventory: [] },
              headers: internal_headers,
              as: :json
          end

          assert_response :unprocessable_entity
          assert_equal({ "error" => "invalid_input" }, JSON.parse(response.body))
          assert_nil @user.reload.player_inventory
        end

        private

        def internal_headers(token = "test-internal-token")
          { "Authorization" => "Bearer #{token}" }
        end

        def with_internal_token_env
          previous_value = ENV["GAME_SERVER_INTERNAL_TOKEN"]
          ENV["GAME_SERVER_INTERNAL_TOKEN"] = "test-internal-token"
          yield
        ensure
          if previous_value.nil?
            ENV.delete("GAME_SERVER_INTERNAL_TOKEN")
          else
            ENV["GAME_SERVER_INTERNAL_TOKEN"] = previous_value
          end
        end
      end
    end
  end
end
