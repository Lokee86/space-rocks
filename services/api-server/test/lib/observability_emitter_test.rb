require "test_helper"
require "json"
require "stringio"

class ObservabilityEmitterTest < ActiveSupport::TestCase
  INSTANCE_ID = "550e8400-e29b-41d4-a716-446655440010"
  EVENT_ID = "550e8400-e29b-41d4-a716-446655440011"
  TRACE_ID = "550e8400-e29b-41d4-a716-446655440012"

  class FailingWriter
    def write(*)
      raise IOError, "disk unavailable"
    end
  end

  def emitter(writer: StringIO.new, service: Observability::ContractGenerated::SERVICE_API_SERVER)
    identity = Observability::ProcessIdentity.new(service_instance_id: INSTANCE_ID, worker_id: "worker-1", pid: 42)
    Observability::Emitter.new(
      identity: identity,
      writer: writer,
      service: service,
      build_version: "test-build",
      environment: "test",
      clock: -> { Time.utc(2026, 7, 14, 1, 2, 3) },
      uuid_generator: -> { EVENT_ID },
      warning_io: StringIO.new
    )
  end

  test "emits only the canonical envelope" do
    writer = StringIO.new
    result = emitter(writer: writer).emit(event: "build_version_loaded", message: "loaded", fields: { "mode" => "test" })

    assert result[:accepted]
    record = JSON.parse(writer.string)
    assert_equal %w[build_version category environment event event_id fields level message pid retention_tier schema_version service service_instance_id timestamp worker_id], record.keys.sort
    assert_equal "api-server", record["service"]
    assert_equal "build_version_loaded", record["event"]
    assert_equal EVENT_ID, record["event_id"]
  end

  test "ordinary emission cannot use the bridge event" do
    result = emitter.emit(event: "log_message", message: "forbidden")
    assert_equal "bridge_event_forbidden", result[:rejection_code]
  end


  test "redacts generated redact fields and discards rejected unsafe content" do
    writer = StringIO.new
    runtime = emitter(writer: writer)
    accepted = runtime.emit(event: "build_version_loaded", fields: { "private_profile" => "sensitive" })
    assert accepted[:accepted]
    assert accepted[:redacted]
    assert_equal "[REDACTED]", JSON.parse(writer.string).dig("fields", "private_profile")

    writer.truncate(0)
    writer.rewind
    rejected = runtime.emit(event: "build_version_loaded", fields: { "password" => "do-not-leak" })
    assert_equal "unsafe_field", rejected[:rejection_code]
    assert_empty writer.string
  end

  test "uses stable rejection codes for trace service UUID and limits" do
    runtime = emitter
    assert_equal "trace_required", runtime.emit(event: "api_request_started")[:rejection_code]
    assert_equal "invalid_uuid", runtime.emit(event: "api_request_started", context: { trace_id: "bad" })[:rejection_code]
    assert_equal "service_not_allowed", runtime.emit(event: "storage_backend_selected")[:rejection_code]
    assert_equal "invalid_field_key", runtime.emit(event: "build_version_loaded", fields: { "Bad-Key" => 1 })[:rejection_code]
    assert_equal "invalid_field_type", runtime.emit(event: "build_version_loaded", fields: { "items" => [] })[:rejection_code]
    assert_equal "null_not_allowed", runtime.emit(event: "build_version_loaded", fields: { "item" => nil })[:rejection_code]
    assert_equal "string_limit_exceeded", runtime.emit(event: "build_version_loaded", fields: { "item" => "x" * 4097 })[:rejection_code]
    fields = 33.times.to_h { |index| ["item_#{index}", index] }
    assert_equal "field_limit_exceeded", runtime.emit(event: "build_version_loaded", fields: fields)[:rejection_code]
  end

  test "writer failure is non-raising and operationally visible" do
    runtime = emitter(writer: FailingWriter.new)
    result = runtime.emit(event: "build_version_loaded")

    assert result[:write_failed]
    assert_equal "write_failed", result[:rejection_code]
    assert_equal 1, runtime.status[:write_failure_count]
    assert_match(/disk unavailable/, runtime.status[:last_write_error])
  end

  test "shared cross-language fixtures have matching outcomes" do
    path = File.expand_path("../../../../shared/contracts/observability/fixtures/emitter_cases.json", __dir__)
    fixture = JSON.parse(File.read(path))
    fixture.fetch("cases").each do |test_case|
      next if test_case.fetch("mode") == "legacy"

      runtime = emitter
      result = runtime.emit(
        event: test_case.fetch("event"),
        message: test_case.fetch("message", ""),
        context: test_case.fetch("context", {}),
        fields: test_case.fetch("fields", {})
      )
      assert_equal test_case.fetch("accepted"), result[:accepted], test_case.fetch("id")
      assert_equal test_case["redacted"], result[:redacted], test_case.fetch("id") if test_case.key?("redacted")
      if test_case.key?("rejection_code")
        assert_equal test_case["rejection_code"], result[:rejection_code], test_case.fetch("id")
      end
    end
  end
end
