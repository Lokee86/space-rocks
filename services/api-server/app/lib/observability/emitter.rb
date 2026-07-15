require "json"
require "securerandom"
require "time"
require_relative "contract_generated"

module Observability
  class Emitter
    CONTEXT_FIELDS = %w[
      trace_id session_id room_id player_id account_id match_id result_id request_id
      route packet_type duration_ms failure_mode reason_code diagnostic_report_id audit_event_id
    ].freeze
    UUID_CONTEXT_FIELDS = %w[trace_id diagnostic_report_id audit_event_id].freeze

    attr_reader :status

    def initialize(identity:, writer:, service: ContractGenerated::SERVICE_API_SERVER,
                   build_version: "development", environment: "development",
                   clock: -> { Time.now }, uuid_generator: -> { SecureRandom.uuid }, warning_io: $stderr,
                   development: false)
      @identity = identity
      @writer = writer
      @service_key = service
      @service = ContractGenerated::SERVICE_DEFINITIONS.fetch(service)
      @build_version = build_version
      @environment = environment
      @clock = clock
      @uuid_generator = uuid_generator
      @warning_io = warning_io
      @development = development
      @last_warning_at = nil
      @status = {
        accepted_count: 0, rejected_count: 0, redacted_count: 0,
        write_failure_count: 0, last_rejection_code: nil, last_write_error: nil
      }
      validate_identity!
    end

    def emit(event:, message: nil, context: {}, fields: {})
      definition = ContractGenerated::EVENT_DEFINITIONS[event]
      return reject(ContractGenerated::REJECTION_UNKNOWN_EVENT) unless definition
      return reject(ContractGenerated::REJECTION_BRIDGE_EVENT_FORBIDDEN) if definition.fetch("bridge_only")

      emit_definition(definition, level: definition.fetch("default_level"),
        category: definition.fetch("category"), message: message, context: context, fields: fields)
    end

    private

    def emit_definition(definition, level:, category:, message:, context:, fields:, at: nil)
      return reject(ContractGenerated::REJECTION_SERVICE_NOT_ALLOWED) unless definition.fetch("services").include?(@service_key)

      normalized_context, context_error = normalize_context(context)
      return reject(*context_error) if context_error
      if definition.fetch("trace_required") && normalized_context["trace_id"].to_s.empty?
        return reject(ContractGenerated::REJECTION_TRACE_REQUIRED, "trace_id")
      end

      safe_fields, redacted, field_error = sanitize_fields(fields)
      return reject(*field_error) if field_error
      if oversized?(message) || oversized?(category)
        return reject(ContractGenerated::REJECTION_STRING_LIMIT_EXCEEDED, "message")
      end

      event_id = @uuid_generator.call.to_s
      return reject(ContractGenerated::REJECTION_INVALID_UUID, "event_id") unless uuid?(event_id)

      record = {
        "timestamp" => (at || @clock.call).utc.iso8601(9),
        "level" => level,
        "event" => definition.fetch("name"),
        "event_id" => event_id,
        "service" => @service.fetch("emitted_name"),
        "environment" => @environment,
        "build_version" => @build_version,
        "schema_version" => ContractGenerated::SCHEMA_VERSION,
        "service_instance_id" => @identity.service_instance_id,
        "category" => category,
        "retention_tier" => definition.fetch("retention_tier")
      }
      record["message"] = message unless message.nil? || message.empty?
      normalized_context.each { |key, value| record[key] = value unless value.nil? || value == "" }
      record["worker_id"] = @identity.worker_id unless @identity.worker_id.to_s.empty?
      record["pid"] = @identity.pid if @identity.pid
      record["fields"] = safe_fields unless safe_fields.empty?
      return reject(ContractGenerated::REJECTION_FIELD_LIMIT_EXCEEDED) if record.size > ContractGenerated::MAX_EVENT_FIELDS

      begin
        payload = JSON.generate(record)
      rescue StandardError
        return reject(ContractGenerated::REJECTION_SERIALIZATION_FAILED)
      end
      return reject(ContractGenerated::REJECTION_EVENT_TOO_LARGE) if payload.bytesize > ContractGenerated::MAX_EVENT_BYTES

      begin
        @writer.write(payload + "\n")
        writer_status = @writer.respond_to?(:status) ? @writer.status : {}
        if writer_status[:degraded]
          return write_failure(writer_status[:last_error] || "writer degraded")
        end
      rescue StandardError => error
        return write_failure("#{error.class}: #{error.message}")
      end

      @status[:accepted_count] += 1
      @status[:redacted_count] += 1 if redacted
      { accepted: true, redacted: redacted, rejection_code: nil, rejected_key: nil, write_failed: false }
    end

    def normalize_context(context)
      return [{}, [ContractGenerated::REJECTION_INVALID_FIELD_TYPE, "context"]] unless context.is_a?(Hash)
      normalized = context.transform_keys(&:to_s)
      unknown = normalized.keys - CONTEXT_FIELDS
      return [{}, [ContractGenerated::REJECTION_UNKNOWN_CONTEXT_FIELD, unknown.first]] unless unknown.empty?

      normalized.each do |key, value|
        return [{}, [ContractGenerated::REJECTION_NULL_NOT_ALLOWED, key]] if value.nil?
        valid = key == "duration_ms" ? value.is_a?(Numeric) && value.finite? : value.is_a?(String)
        return [{}, [ContractGenerated::REJECTION_INVALID_FIELD_TYPE, key]] unless valid
        return [{}, [ContractGenerated::REJECTION_INVALID_UUID, key]] if UUID_CONTEXT_FIELDS.include?(key) && !value.empty? && !uuid?(value)
        return [{}, [ContractGenerated::REJECTION_STRING_LIMIT_EXCEEDED, key]] if value.is_a?(String) && oversized?(value)
      end
      [normalized, nil]
    end

    def sanitize_fields(fields)
      return [{}, false, [ContractGenerated::REJECTION_INVALID_FIELD_TYPE, "fields"]] unless fields.is_a?(Hash)
      return [{}, false, [ContractGenerated::REJECTION_FIELD_LIMIT_EXCEEDED, nil]] if fields.size > ContractGenerated::MAX_FREE_FORM_FIELDS

      safe = {}
      redacted = false
      fields.each do |raw_key, value|
        key = raw_key.to_s
        action, matched, ambiguous = redaction_action(key)
        return [{}, false, [ContractGenerated::REJECTION_UNSAFE_FIELD, key]] if ambiguous || (matched && action == "reject")
        if matched && action == "redact"
          safe[key] = ContractGenerated::REDACTION_REPLACEMENT_MARKER
          redacted = true
          next
        end
        return [{}, false, [ContractGenerated::REJECTION_INVALID_FIELD_KEY, key]] unless ContractGenerated::FREE_FORM_KEY_REGEX.match?(key)
        return [{}, false, [ContractGenerated::REJECTION_NULL_NOT_ALLOWED, key]] if value.nil?
        unless value.is_a?(String) || value.is_a?(Numeric) || value == true || value == false
          return [{}, false, [ContractGenerated::REJECTION_INVALID_FIELD_TYPE, key]]
        end
        if value.is_a?(Numeric) && !value.finite?
          return [{}, false, [ContractGenerated::REJECTION_INVALID_FIELD_TYPE, key]]
        end
        return [{}, false, [ContractGenerated::REJECTION_STRING_LIMIT_EXCEEDED, key]] if value.is_a?(String) && value.bytesize > ContractGenerated::MAX_FREE_FORM_VALUE_BYTES
        safe[key] = value
      end
      [safe, redacted, nil]
    end

    def redaction_action(key)
      candidate = ContractGenerated::REDACTION_CASE_SENSITIVE ? key : key.downcase
      actions = []
      ContractGenerated::REDACTION_EXACT_RULES.each do |rule|
        actions << rule.fetch("action") if rule.fetch("matches").include?(candidate)
      end
      ContractGenerated::REDACTION_FRAGMENT_RULES.each do |rule|
        actions << rule.fetch("action") if rule.fetch("matches").any? { |fragment| candidate.include?(fragment) }
      end
      actions.uniq!
      return [nil, false, false] if actions.empty?
      return [ContractGenerated::REDACTION_AMBIGUOUS_MATCH_ACTION, true, true] if actions.size > 1
      [actions.first, true, false]
    end

    def reject(code, key = nil)
      @status[:rejected_count] += 1
      @status[:last_rejection_code] = code
      warn_rejection(code, key)
      { accepted: false, redacted: false, rejection_code: code, rejected_key: key, write_failed: false }
    end

    def write_failure(error)
      @status[:write_failure_count] += 1
      @status[:last_rejection_code] = ContractGenerated::REJECTION_WRITE_FAILED
      @status[:last_write_error] = error.to_s
      warn_rejection(ContractGenerated::REJECTION_WRITE_FAILED, nil)
      { accepted: false, redacted: false, rejection_code: ContractGenerated::REJECTION_WRITE_FAILED, rejected_key: nil, write_failed: true }
    end

    def warn_rejection(code, key)
      now = @clock.call
      return if @last_warning_at && now - @last_warning_at < 5
      @last_warning_at = now
      suffix = @development && key ? " key=#{key}" : ""
      @warning_io&.puts("observability event rejected service=#{@service.fetch("emitted_name")} code=#{code}#{suffix}")
    rescue StandardError
      nil
    end

    def validate_identity!
      raise ArgumentError, "service instance ID must be a UUID" unless uuid?(@identity.service_instance_id)
      raise ArgumentError, "environment and build version are required" if @environment.to_s.empty? || @build_version.to_s.empty?
    end

    def uuid?(value)
      ContractGenerated::UUID_REGEX.match?(value.to_s)
    end

    def oversized?(value)
      value && value.to_s.bytesize > ContractGenerated::MAX_STRING_BYTES
    end
  end
end
