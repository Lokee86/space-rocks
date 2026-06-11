require "test_helper"

class OpenapiContractTest < ActiveSupport::TestCase
  test "openapi contract parses" do
    definition = OpenapiFirst.load(Rails.root.join("../../shared/contracts/http/openapi.yaml").expand_path)

    assert definition.present?
  end
end
