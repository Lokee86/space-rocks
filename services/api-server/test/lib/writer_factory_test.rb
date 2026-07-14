require "test_helper"

class WriterFactoryTest < ActiveSupport::TestCase
  test "derives unique active paths for workers and pids" do
    configuration = Observability::ConfigurationFactory.from_env("API_OBSERVABILITY_LOG_ROOT" => "tmp/observability")
    paths = [
      ["worker-1", 10],
      ["worker-2", 10],
      ["worker-1", 11]
    ].map do |worker_id, pid|
      identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: worker_id, pid: pid)
      Observability::WriterFactory.new(configuration, identity, writer: ->(path) { path }).call
    end

    assert_equal 3, paths.uniq.length
    assert paths.all? { |path| path.end_with?(".jsonl.open") }
  end

  test "delegates writer construction with the derived path" do
    configuration = Observability::ConfigurationFactory.from_env("API_OBSERVABILITY_LOG_ROOT" => "tmp/observability")
    identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: "worker/1", pid: 10)
    received = nil

    result = Observability::WriterFactory.new(configuration, identity, writer: ->(path) { received = path; :writer }).call

    assert_equal :writer, result
    assert_equal "tmp/observability/active/api-server-api-a-worker_1-pid-10.jsonl.open", received
  end
end