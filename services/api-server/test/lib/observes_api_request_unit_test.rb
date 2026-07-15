require "minitest/test"
require "minitest/autorun"
require "active_support"
require "active_support/core_ext/object/blank"
require_relative "../../app/lib/observability/contract_generated"
require_relative "../../app/lib/observability/puma_hooks"
require_relative "../../app/controllers/concerns/observes_api_request"

class ObservesApiRequestUnitTest < Minitest::Test
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

  class Harness
    class << self
      def around_action(*)
      end
    end

    include ObservesApiRequest

    attr_reader :request, :response

    def initialize(request_id: "request-1", trace_id: nil, controller_path: "api/auth/sessions", action_name: "create")
      @request = Request.new(
        request_id: request_id,
        request_method: "POST",
        headers: trace_id ? { "X-Trace-ID" => trace_id } : {}
      )
      @response = Response.new
      @controller_path = controller_path
      @action_name = action_name
    end

    def controller_path
      @controller_path
    end

    def action_name
      @action_name
    end

    def run
      send(:observe_api_request) { yield }
    end
  end

  TRACE_ID = "550e8400-e29b-41d4-a716-446655440012"

  def setup
    @puma_hooks_singleton = Observability::PumaHooks.singleton_class
    @puma_hooks_original_emit = Observability::PumaHooks.method(:emit)
    Observability::PumaHooks.instance_variable_set(:@test_events, [])
    Observability::PumaHooks.instance_variable_set(:@test_raise_errors, false)
    Observability::PumaHooks.instance_variable_set(:@test_reject_events, [])
    @puma_hooks_singleton.define_method(:emit) do |event:, message: nil, context: {}, fields: {}|
      raise IOError, "writer unavailable" if instance_variable_get(:@test_raise_errors)
      if instance_variable_get(:@test_reject_events)&.include?(event)
        next({ accepted: false, rejection_code: "test_rejected" })
      end

      events = instance_variable_get(:@test_events)
      events << { event: event, message: message, context: context, fields: fields }
      { accepted: true }
    end
  end

  def teardown
    @puma_hooks_singleton.define_method(:emit, @puma_hooks_original_emit)
    %i[@test_events @test_raise_errors @test_reject_events].each do |name|
      Observability::PumaHooks.remove_instance_variable(name) if Observability::PumaHooks.instance_variable_defined?(name)
    end
  end

  def test_preserves_a_valid_trace_returns_it_and_emits_one_start_and_completion_with_duration
    harness = Harness.new(trace_id: TRACE_ID)
    harness.run { harness.response.status = 201 }

    assert_equal TRACE_ID, harness.response.headers["X-Trace-ID"]
    assert_equal ["api_request_started", "api_request_completed"], event_names
    completed = test_events.last
    assert_equal "api/auth/sessions#create", completed[:context]["route"]
    assert_equal "request-1", completed[:context]["request_id"]
    assert_equal TRACE_ID, completed[:context]["trace_id"]
    assert_kind_of Numeric, completed[:context]["duration_ms"]
    assert_equal 201, completed[:fields]["status"]
    assert_equal "POST", completed[:fields]["http_method"]
  end

  def test_generates_a_trace_when_the_incoming_value_is_missing_or_invalid
    [nil, "not-a-uuid"].each do |trace_id|
      harness = Harness.new(trace_id: trace_id)
      harness.run { }
      resolved = harness.response.headers["X-Trace-ID"]
      assert_match Observability::ContractGenerated::UUID_REGEX, resolved
      refute_equal trace_id, resolved
    end
  end

  def test_excludes_health_endpoints_while_still_returning_the_resolved_trace
    harness = Harness.new(controller_path: "rails/health", action_name: "show")
    harness.run { }

    assert_match Observability::ContractGenerated::UUID_REGEX, harness.response.headers["X-Trace-ID"]
    assert_empty test_events
  end

  def test_emits_one_request_failure_for_an_unhandled_exception
    harness = Harness.new
    assert_raises(RuntimeError) { harness.run { raise "boom" } }

    assert_equal ["api_request_started", "api_request_failed"], event_names
    assert_equal "unhandled_exception", test_events.last[:context]["failure_mode"]
  end

  def test_does_not_add_request_completion_after_a_specific_workflow_failure
    harness = Harness.new
    harness.run do
      harness.send(:emit_api_event, "auth_failed", context: { "reason_code" => "invalid_token" }, specific_failure: true, status: :unauthorized)
      harness.response.status = 401
    end

    assert_equal ["api_request_started", "auth_failed"], event_names
  end

  def test_rejected_specific_failure_does_not_suppress_the_fallback_request_event
    Observability::PumaHooks.instance_variable_set(:@test_reject_events, ["auth_failed"])
    harness = Harness.new
    harness.run do
      harness.send(:emit_api_event, "auth_failed", context: { "reason_code" => "invalid_token" }, specific_failure: true, status: :unauthorized)
      harness.response.status = 401
    end

    assert_equal ["api_request_started", "api_request_completed"], event_names
    assert_equal 401, test_events.last[:fields]["status"]
  end

  private

  def test_events
    Observability::PumaHooks.instance_variable_get(:@test_events)
  end

  def event_names
    test_events.map { |event| event[:event] }
  end
end
