require_relative "configuration"
require_relative "contract_generated"

module Observability
  module ConfigurationFactory
    module_function

    DEFAULTS = {
      enabled: false,
      log_root: "log/observability",
      service_instance_id: "api-server-local",
      build_version: "development",
      environment: "development",
      segment_bytes: 50 * 1024 * 1024,
      segment_age: ContractGenerated::FILE_LOGGING_MAX_ACTIVE_SEGMENT_AGE_SECONDS,
      retention_age: ContractGenerated::RETENTION_DEFAULT_AGE_SECONDS_OPERATIONAL,
      retention_bytes: 500 * 1024 * 1024,
      compression: ContractGenerated::FILE_LOGGING_COMPRESSION_ENABLED
    }.freeze

    def from_env(env = ENV)
      enabled = boolean(env.fetch("API_OBSERVABILITY_ENABLED", DEFAULTS[:enabled]))
      service_instance_id = text(env, "API_SERVICE_INSTANCE_ID", DEFAULTS[:service_instance_id])
      if enabled && !ContractGenerated::UUID_REGEX.match?(service_instance_id)
        raise ArgumentError, "API_SERVICE_INSTANCE_ID must be a UUID when observability is enabled"
      end

      Configuration.new(
        enabled: enabled,
        log_root: path(env, "API_OBSERVABILITY_LOG_ROOT", DEFAULTS[:log_root]),
        service_instance_id: service_instance_id,
        build_version: text(env, "BUILD_VERSION", DEFAULTS[:build_version]),
        environment: text(env, "RAILS_ENV", DEFAULTS[:environment]),
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

    def path(env, key, default)
      value = text(env, key, default)
      raise ArgumentError, "#{key} contains a null byte" if value.include?("\0")

      value
    end

    def text(env, key, default)
      value = env.fetch(key, default).to_s
      raise ArgumentError, "#{key} must not be blank" if value.strip.empty?

      value
    end

    def parse_duration(value)
      match = /\A(\d+)\s*(s|m|h|d)\z/i.match(value)
      raise ArgumentError unless match

      Integer(match[1]) * { "s" => 1, "m" => 60, "h" => 3600, "d" => 86_400 }.fetch(match[2].downcase)
    end
    private_class_method :parse_duration
  end
end