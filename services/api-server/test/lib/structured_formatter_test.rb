require "test_helper"
require "json"
require "stringio"

class StructuredFormatterTest < ActiveSupport::TestCase
  UUID = "550e8400-e29b-41d4-a716-446655440010"

  class Sink
    attr_reader :records

    def initialize
      @records = []
    end

    def write(record)
      @records << record
    end
  end

  def formatter(sink)
    identity = Observability::ProcessIdentity.new(service_instance_id: UUID, worker_id: "worker-1", pid: 42)
    emitter = Observability::Emitter.new(
      identity: identity,
      writer: sink,
      build_version: "build-1",
      environment: "test",
      uuid_generator: -> { "550e8400-e29b-41d4-a716-446655440011" },
      warning_io: StringIO.new
    )
    Observability::StructuredFormatter.new(emitter)
  end

  test "bridges one canonical JSON object per line with stable context" do
    sink = Sink.new
    result = formatter(sink).call("INFO", Time.utc(2026, 7, 13, 12, 30, 0), nil, { "event" => "started" })

    assert result[:accepted]
    assert_equal "\n", sink.records.sole[-1]
    payload = JSON.parse(sink.records.sole)
    assert_equal "2026-07-13T12:30:00.000000000Z", payload["timestamp"]
    assert_equal "info", payload["level"]
    assert_equal "log_message", payload["event"]
    assert_equal "api-server", payload["service"]
    assert_equal "operational", payload["retention_tier"]
    assert_equal UUID, payload["service_instance_id"]
    assert_equal "worker-1", payload["worker_id"]
    assert_equal 42, payload["pid"]
    assert_equal "started", payload.dig("fields", "legacy_event")
  end

  test "keeps bounded exception metadata without backtrace content" do
    sink = Sink.new
    error = RuntimeError.new("broken")
    error.set_backtrace(["app.rb:3"])
    result = formatter(sink).call("ERROR", Time.utc(2026, 7, 13), nil, error)

    assert result[:accepted]
    payload = JSON.parse(sink.records.sole)
    assert_equal "broken", payload["message"]
    assert_equal "RuntimeError", payload.dig("fields", "error_class")
    assert_not_includes sink.records.sole, "app.rb:3"
  end

  test "normalizes legacy container fields to canonical scalars" do
    sink = Sink.new
    result = formatter(sink).call(
      "INFO", Time.utc(2026, 7, 13), nil,
      { "event" => "channels_ready", "channels" => ["reliable", "unreliable"] }
    )

    assert result[:accepted]
    assert_equal '["reliable","unreliable"]', JSON.parse(sink.records.sole).dig("fields", "channels")
  end
end
