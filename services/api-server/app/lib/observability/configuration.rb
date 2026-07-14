module Observability
  Configuration = Data.define(
    :enabled,
    :log_root,
    :service_instance_id,
    :segment_bytes,
    :segment_age,
    :retention_age,
    :retention_bytes,
    :compression
  )

  module ConfigurationFactory
    module_function

    DEFAULTS = {
      enabled: false,
      log_root: "log/observability",
      service_instance_id: "api-server-local",
      segment_bytes: 50 * 1024 * 1024,
      segment_age: 86_400,
      retention_age: 1_209_600,
      retention_bytes: 500 * 1024 * 1024,
      compression: true
    }.freeze

    def from_env(env = ENV)
      Configuration.new(
        enabled: boolean(env.fetch("API_OBSERVABILITY_ENABLED", DEFAULTS[:enabled])),
        log_root: env.fetch("API_OBSERVABILITY_LOG_ROOT", DEFAULTS[:log_root]),
        service_instance_id: env.fetch("API_SERVICE_INSTANCE_ID", DEFAULTS[:service_instance_id]),
        segment_bytes: positive_integer(env, "API_OBSERVABILITY_SEGMENT_BYTES", DEFAULTS[:segment_bytes]),
        segment_age: duration(env, "API_OBSERVABILITY_SEGMENT_AGE", DEFAULTS[:segment_age]),
        retention_age: duration(env, "API_OBSERVABILITY_RETENTION_AGE", DEFAULTS[:retention_age]),
        retention_bytes: positive_integer(env, "API_OBSERVABILITY_RETENTION_BYTES", DEFAULTS[:retention_bytes]),
        compression: boolean(env.fetch("API_OBSERVABILITY_COMPRESSION", DEFAULTS[:compression]))
      )
    end

    def boolean(value)
      return value if value == true || value == false

      case value.to_s.downcase
      when "1", "true", "yes", "on" then true
      when "0", "false", "no", "off" then false
      else raise ArgumentError, "invalid boolean: #{value.inspect}"
      end
    end

    def positive_integer(env, key, default)
      value = Integer(env.fetch(key, default))
      raise ArgumentError, "#{key} must be positive" unless value.positive?

      value
    rescue ArgumentError, TypeError
      raise ArgumentError, "#{key} must be a positive integer"
    end

    def duration(env, key, default)
      raw = env.fetch(key, default)
      value = raw.is_a?(Numeric) ? raw.to_i : parse_duration(raw.to_s)
      raise ArgumentError, "#{key} must be positive" unless value.positive?

      value
    rescue ArgumentError, TypeError
      raise ArgumentError, "#{key} must be a positive duration"
    end

    def parse_duration(value)
      match = /\A(\d+)\s*(s|m|h|d)\z/i.match(value)
      raise ArgumentError unless match

      Integer(match[1]) * { "s" => 1, "m" => 60, "h" => 3600, "d" => 86_400 }.fetch(match[2].downcase)
    end
    private_class_method :parse_duration
  end
end