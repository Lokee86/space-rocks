require "test_helper"

class HealthControllerTest < ActionDispatch::IntegrationTest
  test "GET /health returns static status" do
    get "/health"

    assert_response :success

    body = JSON.parse(response.body)
    assert_equal "ok", body["status"]
    assert_equal "space-rocks-api", body["service"]
  end
end
