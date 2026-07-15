require "json"
require "time"

module Observability
  class StructuredFormatter
    def initialize(identity, build_version: "development", environment: "development")
      @identity = identity
      @build_version = build_version
      @environment = environment
    end

    def call(severity, time, _progname, message)
      payload = {
        timestamp: time.utc.iso8601(3),
        level: severity,
        service: "api-server",
        environment: @environment,
        build_version: @build_version,
        service_instance_id: @identity.service_instance_id,
        worker_id: @identity.worker_id,
        pid: @identity.pid,
        message: message_value(message)
      }
      payload[:exception] = exception_value(message) if message.is_a?(Exception)

      JSON.generate(payload) + "\n"
    end

    private

    def message_value(message)
      return message.message if message.is_a?(Exception)
      return message if message.is_a?(String) || message.is_a?(Hash)

      message.to_s
    end

    def exception_value(exception)
      {
        class: exception.class.name,
        message: exception.message,
        backtrace: exception.backtrace
      }
    end
  end
end