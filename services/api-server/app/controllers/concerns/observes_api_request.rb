require "securerandom"

module ObservesApiRequest
  extend ActiveSupport::Concern

  HEALTH_CONTROLLER_PATHS = %w[health rails/health].freeze
  TRACE_HEADER = "X-Trace-ID"
  RESPONSE_TRACE_HEADER = "X-Trace-ID"
  HTTP_STATUS_CODES = {
    accepted: 202,
    bad_gateway: 502,
    bad_request: 400,
    not_found: 404,
    ok: 200,
    unauthorized: 401,
    unprocessable_entity: 422
  }.freeze

  included do
    around_action :observe_api_request
  end

  private

  def observe_api_request
    establish_api_request_context
    return yield if api_health_request?

    emit_api_event(
      "api_request_started",
      fields: request_event_fields,
      include_duration: false
    )

    error = nil
    begin
      yield
    rescue Exception => exception
      error = exception
      raise
    ensure
      if error
        unless api_specific_failure_emitted?
          emit_api_event(
            "api_request_failed",
            context: { "failure_mode" => "unhandled_exception" },
            fields: request_event_fields(status: 500),
            specific_failure: true
          )
        end
      elsif !api_specific_failure_emitted?
        emit_api_event(
          "api_request_completed",
          fields: request_event_fields
        )
      end
    end
  end

  def establish_api_request_context
    @api_request_started_at = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    @api_request_id = request.request_id.to_s.presence || SecureRandom.uuid
    @api_trace_id = resolved_trace_id
    @api_route = "#{controller_path}##{action_name}"
    response.set_header(RESPONSE_TRACE_HEADER, @api_trace_id)
  end

  def resolved_trace_id
    incoming_trace_id = request.headers[TRACE_HEADER].to_s
    return incoming_trace_id if Observability::ContractGenerated::UUID_REGEX.match?(incoming_trace_id)

    SecureRandom.uuid
  end

  def api_health_request?
    HEALTH_CONTROLLER_PATHS.include?(controller_path)
  end

  def request_event_fields(status: nil)
    fields = {
      "http_method" => request.request_method,
      "status" => status || response.status.to_i
    }
    fields
  end

  def emit_api_event(event, context: {}, fields: {}, message: nil, specific_failure: false, include_duration: true, status: nil)
    emitted_fields = fields.dup
    emitted_fields["status"] = api_status_code(status) unless status.nil?
    result = Observability::PumaHooks.emit(
      event: event,
      message: message,
      context: api_request_context(include_duration: include_duration).merge(context).compact,
      fields: emitted_fields
    )
    @api_specific_failure_emitted = true if specific_failure && result.is_a?(Hash) && result[:accepted]
    result
  rescue StandardError
    nil
  end

  def api_request_context(include_duration: true)
    context = {
      "request_id" => @api_request_id,
      "trace_id" => @api_trace_id,
      "route" => @api_route,
      "account_id" => @api_account_id
    }
    context["duration_ms"] = request_duration_ms if include_duration
    context.compact
  end

  def request_duration_ms
    return nil unless @api_request_started_at

    ((Process.clock_gettime(Process::CLOCK_MONOTONIC) - @api_request_started_at) * 1000.0).round(3)
  end

  def set_api_account_id!(account_id)
    @api_account_id = account_id.to_s.presence
  end

  def mark_api_specific_failure!
    @api_specific_failure_emitted = true
  end

  def api_specific_failure_emitted?
    @api_specific_failure_emitted == true
  end

  def api_reason_code(value, fallback)
    candidate = value.to_s
    return candidate if /\A[a-z][a-z0-9_]{0,63}\z/.match?(candidate)

    fallback
  end

  def api_status_code(status)
    return status if status.is_a?(Numeric)

    HTTP_STATUS_CODES.fetch(status.to_sym)
  end
end
