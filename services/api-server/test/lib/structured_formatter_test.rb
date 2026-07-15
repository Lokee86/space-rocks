require "test_helper"
require "json"
require "time"

class StructuredFormatterTest < ActiveSupport::TestCase
  test "formats one JSON object per line with stable context" do
    identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 42)
    time = Time.utc(2026, 7, 13, 12, 30, 0)
    output = Observability::StructuredFormatter.new(identity, build_version: "build-1", environment: "test").call("INFO", time, nil, { "event" => "started" })

    assert_equal "\n", output[-1]
    assert_equal({
      "timestamp" => "2026-07-13T12:30:00.000Z",
      "level" => "INFO",
      "service" => "api-server",
      "environment" => "test",
      "build_version" => "build-1",
      "service_instance_id" => "api-a",
      "worker_id" => "worker-1",
      "pid" => 42,
      "message" => { "event" => "started" }
    }, JSON.parse(output))
  end

  test "includes exception metadata" do
    error = RuntimeError.new("broken")
    error.set_backtrace(["app.rb:3"])
    identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: "single-process", pid: 42)
    output = Observability::StructuredFormatter.new(identity).call("ERROR", Time.utc(2026, 7, 13), nil, error)

    assert_equal({ "class" => "RuntimeError", "message" => "broken", "backtrace" => ["app.rb:3"] }, JSON.parse(output)["exception"])
    assert_equal "broken", JSON.parse(output)["message"]
  end
end