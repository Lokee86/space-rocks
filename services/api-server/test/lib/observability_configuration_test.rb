require "test_helper"

class ObservabilityConfigurationTest < ActiveSupport::TestCase
  Factory = Observability::ConfigurationFactory
  UUID = "550e8400-e29b-41d4-a716-446655440010"

  test "uses generated policy defaults" do
    configuration = Factory.from_env({})

    assert_not configuration.enabled
    assert_equal "log/observability", configuration.log_root
    assert_equal 50 * 1024 * 1024, configuration.segment_bytes
    assert_equal Observability::ContractGenerated::FILE_LOGGING_MAX_ACTIVE_SEGMENT_AGE_SECONDS, configuration.segment_age
    assert_equal Observability::ContractGenerated::RETENTION_DEFAULT_AGE_SECONDS_OPERATIONAL, configuration.retention_age
    assert_equal Observability::ContractGenerated::FILE_LOGGING_COMPRESSION_ENABLED, configuration.compression
  end

  test "loads environment overrides" do
    configuration = Factory.from_env(
      "API_OBSERVABILITY_ENABLED" => "true",
      "API_OBSERVABILITY_LOG_ROOT" => "tmp/logs",
      "API_SERVICE_INSTANCE_ID" => UUID,
      "BUILD_VERSION" => "build-1",
      "RAILS_ENV" => "staging",
      "API_OBSERVABILITY_SEGMENT_BYTES" => "1024",
      "API_OBSERVABILITY_SEGMENT_AGE" => "2h",
      "API_OBSERVABILITY_RETENTION_AGE" => "3d",
      "API_OBSERVABILITY_RETENTION_BYTES" => "4096",
      "API_OBSERVABILITY_COMPRESSION" => "false"
    )

    assert_equal [true, "tmp/logs", UUID, "build-1", "staging", 1024, 7200, 259_200, 4096, false], configuration.to_h.values
  end

  test "rejects invalid paths identity numeric and duration values" do
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_LOG_ROOT" => " ") }
    assert_raises(ArgumentError) { Factory.from_env("API_SERVICE_INSTANCE_ID" => "") }
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_ENABLED" => "true", "API_SERVICE_INSTANCE_ID" => "not-a-uuid") }
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_SEGMENT_BYTES" => "nope") }
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_SEGMENT_AGE" => "soon") }
    assert_raises(ArgumentError) { Factory.from_env("API_OBSERVABILITY_RETENTION_BYTES" => "0") }
  end

  test "configuration is immutable" do
    assert_raises(NoMethodError) { Factory.from_env.enabled = true }
  end
end