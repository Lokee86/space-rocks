require "json"
require "minitest/test"
require "minitest/autorun"
require "stringio"
require "active_support"
require "active_support/core_ext/object/blank"
require_relative "../../app/lib/observability/contract_generated"
require_relative "../../app/lib/observability/puma_hooks"
require_relative "../../app/controllers/concerns/observes_api_request"
require_relative "../../app/lib/observability/emitter"

class ObservabilityRealEmitterUnitTest < Minitest::Test
  Identity = Struct.new(:service_instance_id, :worker_id, :pid, keyword_init: true)
  Request = Struct.new(:request_id, :request_method, :headers, keyword_init: true)

  class Response
    attr_reader :headers
    attr_accessor :status

    def initialize(status: 200)
      @status = status
      @headers = {}
    end

    def set_header(name, value)
      @headers[name] = value
    end
  end

  class ControllerHarness
    class << self
      def around_action(*)
      end
    end

    include ObservesApiRequest

    attr_reader :request, :response

    def initialize
      @request = Request.new(request_id: "request-1", request_method: "POST", headers: {})
      @response = Response.new
    end

    def controller_path
      "api/test"
    end

    def action_name
      "create"
    end

    def run
      send(:observe_api_request) { yield }
    end
  end

  TRACE_ID = "550e8400-e29b-41d4-a716-446655440012"
  SERVICE_INSTANCE_ID = "550e8400-e29b-41d4-a716-446655440013"

  def test_controller_workflow_events_reach_the_real_emitter_as_canonical_jsonl
    writer = StringIO.new
    harness = ControllerHarness.new

    with_real_emitter(build_emitter(writer)) do
      harness.run do
      harness.send(:emit_api_event, "api_unauthorized", context: { "reason_code" => "authorization_missing" }, specific_failure: true, status: :unauthorized)
      harness.send(:emit_api_event, "auth_failed", context: { "reason_code" => "invalid_credentials" }, fields: { "flow" => "login" }, specific_failure: true, status: :unauthorized)
      harness.send(:emit_api_event, "api_request_failed", context: { "failure_mode" => "stats_serialization_failure" }, specific_failure: true, status: 500)
      harness.send(:emit_api_event, "match_result_report_succeeded", context: { "match_id" => "match-1", "result_id" => "result-1" }, fields: { "status" => 200 }, specific_failure: false)
      harness.response.status = 401
      end
    end

    records = writer.string.lines.map { |line| JSON.parse(line) }
    assert_equal ["api_request_started", "api_unauthorized", "auth_failed", "api_request_failed", "match_result_report_succeeded"], records.map { |record| record["event"] }
    assert_equal "authorization_missing", records[1]["reason_code"]
    assert_equal "invalid_credentials", records[2]["reason_code"]
    assert_equal "stats_serialization_failure", records[3]["failure_mode"]
    assert_equal "result-1", records[4]["result_id"]
    assert_equal 401, records[1]["fields"]["status"]
    assert_equal 500, records[3]["fields"]["status"]
    records.each do |record|
      refute record.fetch("fields", {}).key?("reason_code")
      refute record.fetch("fields", {}).key?("failure_mode")
      refute record.fetch("fields", {}).key?("result_id")
    end
    refute_includes writer.string, "log_message"
  end

  def test_api_match_result_event_is_written_as_canonical_jsonl
    writer = StringIO.new
    result = build_emitter(writer).emit(
      event: "match_result_report_failed",
      context: {
        "trace_id" => TRACE_ID,
        "request_id" => "request-1",
        "route" => "internal/player_data/match_results#create",
        "account_id" => "acct-1",
        "match_id" => "match-1",
        "result_id" => "result-duplicate",
        "failure_mode" => "validation",
        "reason_code" => "invalid_input"
      },
      fields: { "failure_stage" => "validation", "status" => 422 }
    )

    assert result[:accepted]
    record = JSON.parse(writer.string)
    assert_equal "match_result_report_failed", record["event"]
    assert_equal "result-duplicate", record["result_id"]
    assert_equal "validation", record["failure_mode"]
    assert_equal "invalid_input", record["reason_code"]
    assert_equal 422, record["fields"]["status"]
    refute record["fields"].key?("result_id")
    refute record["fields"].key?("failure_mode")
    refute record["fields"].key?("reason_code")
    refute_includes writer.string, "log_message"
  end

  def test_representative_api_workflow_events_are_accepted_by_the_real_emitter
    writer = StringIO.new
    emitter = build_emitter(writer)
    samples = [
      ["api_request_completed", {}, { "status" => 200, "http_method" => "GET" }],
      ["api_unauthorized", { "reason_code" => "authorization_missing" }, { "status" => 401 }],
      ["auth_failed", { "reason_code" => "invalid_credentials" }, { "flow" => "login", "status" => 401 }],
      ["api_request_failed", { "failure_mode" => "stats_serialization_failure" }, { "status" => 500 }],
      ["match_result_duplicate_suppressed", { "result_id" => "result-duplicate", "match_id" => "match-1" }, { "duplicate" => true, "status" => 200 }]
    ]

    results = samples.map do |event, context, fields|
      emitter.emit(
        event: event,
        context: { "trace_id" => TRACE_ID, "request_id" => "request-1", "route" => "api/test#action" }.merge(context),
        fields: fields
      )
    end

    assert results.all? { |result| result[:accepted] }
    records = writer.string.lines.map { |line| JSON.parse(line) }
    assert_equal samples.map(&:first), records.map { |record| record["event"] }
    records.each do |record|
      refute record.fetch("fields", {}).key?("reason_code")
      refute record.fetch("fields", {}).key?("failure_mode")
      refute record.fetch("fields", {}).key?("result_id")
    end
  end

  private

  def with_real_emitter(emitter)
    singleton_class = Observability::PumaHooks.singleton_class
    original_emit = Observability::PumaHooks.method(:emit)
    singleton_class.define_method(:emit) { |**kwargs| emitter.emit(**kwargs) }
    yield
  ensure
    singleton_class.define_method(:emit, original_emit)
  end

  def build_emitter(writer)
    Observability::Emitter.new(
      identity: Identity.new(service_instance_id: SERVICE_INSTANCE_ID, worker_id: "worker-1", pid: 42),
      writer: writer,
      clock: -> { Time.utc(2026, 7, 15, 12, 0, 0) },
      uuid_generator: -> { "550e8400-e29b-41d4-a716-446655440014" }
    )
  end
end
