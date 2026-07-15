require_relative "contract_generated"
require_relative "emitter"

module Observability
  class StructuredFormatter
    LEVELS = {
      "DEBUG" => ContractGenerated::LEVEL_DEBUG,
      "INFO" => ContractGenerated::LEVEL_INFO,
      "WARN" => ContractGenerated::LEVEL_WARN,
      "ERROR" => ContractGenerated::LEVEL_ERROR,
      "FATAL" => ContractGenerated::LEVEL_CRITICAL,
      "UNKNOWN" => ContractGenerated::LEVEL_INFO
    }.freeze

    def initialize(emitter)
      @emitter = emitter
    end

    def call(severity, time, _program_name, message)
      normalized_message, fields, legacy_event = normalize_message(message)
      @emitter.emit_legacy(
        level: LEVELS.fetch(severity.to_s.upcase, ContractGenerated::LEVEL_INFO),
        category: "rails",
        message: normalized_message,
        fields: fields,
        legacy_event: legacy_event,
        at: time
      )
    end

    private

    def normalize_message(message)
      case message
      when Exception
        [message.message, { "error_class" => message.class.name }, nil]
      when Hash
        values = message.transform_keys(&:to_s)
        text = values.delete("message")
        legacy_event = values.delete("event")
        [text&.to_s, normalize_fields(values), legacy_event&.to_s]
      else
        [message.to_s, {}, nil]
      end
    end

    def normalize_fields(fields)
      fields.transform_values do |value|
        case value
        when Hash, Array
          JSON.generate(value)
        when NilClass
          "null"
        when String, TrueClass, FalseClass, Integer, Float
          value
        else
          value.to_s
        end
      end
    end
  end

  class CanonicalLogger
    SEVERITIES = %w[DEBUG INFO WARN ERROR FATAL UNKNOWN].freeze
    attr_accessor :level

    def initialize(formatter)
      @formatter = formatter
      @level = ::Logger::DEBUG
    end

    def add(severity, message = nil, program_name = nil)
      return true if severity.to_i < @level.to_i

      message = yield if message.nil? && block_given?
      @formatter.call(SEVERITIES.fetch(severity.to_i, "UNKNOWN"), Time.now, program_name, message)
      true
    end

    def debug(message = nil, &block) = add(::Logger::DEBUG, message, &block)
    def info(message = nil, &block) = add(::Logger::INFO, message, &block)
    def warn(message = nil, &block) = add(::Logger::WARN, message, &block)
    def error(message = nil, &block) = add(::Logger::ERROR, message, &block)
    def fatal(message = nil, &block) = add(::Logger::FATAL, message, &block)
    def unknown(message = nil, &block) = add(::Logger::UNKNOWN, message, &block)

    def <<(message)
      info(message)
      self
    end

    def close = self
  end
end
