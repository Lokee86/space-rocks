require "test_helper"

class ObservabilityConfigurationTest < ActiveSupport::TestCase
  Factory = Observability::ConfigurationFactory

  test "uses dormant local defaults" do
    configuration = Factory.from_env({})

    assert_not configuration.enabled
    assert_equal "log/observability", configuration.log_root
    assert_equal 50 * 1024 * 1024, configuration.segment_bytes
    assert_equal 86_400, configuration.segment_age
    assert_equal 14 * 86_400, configuration.retention_age
    assert configuration.compression
  end

  test "loads environment overrides" do
    configuration = Factory.from_env(
      "API_OBSERVABILITY_ENABLED" => "true",
      "API_OBSERVABILITY_LOG_ROOT" => "tmp/logs",
      "API_SERVICE_INSTANCE_ID" => "api-a",
      "API_OBSERVABILITY_SEGMENT_BYTES" => "1024",
      "API_OBSERVABILITY_SEGMENT_AGE" => "2h",
      "API_OBSERVABILITY_RETENTION_AGE" => "3d",
      "API_OBSERVABILITY_RETENTION_BYTES" => "4096",
      "API_OBSERVABILITY_COMPRESSION" => "false"
    )

    assert_equal [true, "tmp/logs", "api-a", 1024, 7200, 259_200, 4096, false], configuration.to_h.values
  end

  test "rejects invalid numeric and duration values" do
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_SEGMENT_BYTES" => "nope") }
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_SEGMENT_AGE" => "soon") }
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_RETENTION_BYTES" => "0") }
  end

  test "configuration is immutable" do
    assert_raises(NoMethodError) { Factory.from_env.enabled = true }
  end
end