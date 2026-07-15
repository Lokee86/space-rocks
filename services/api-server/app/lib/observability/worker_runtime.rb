require "logger"
require "thread"
require_relative "archive_store"
require_relative "configuration_factory"
require_relative "emitter"
require_relative "process_identity"
require_relative "structured_formatter"
require_relative "writer_factory"

module Observability
  class WorkerRuntime
    def initialize(env: ENV, logger_provider: -> { Rails.logger }, archive_builder: nil, writer_builder: nil)
      @env = env
      @logger_provider = logger_provider
      @archive_builder = archive_builder || ->(configuration, identity) { ArchiveStore.new(configuration, identity) }
      @writer_builder = writer_builder || lambda do |configuration, identity, archive_store|
        WriterFactory.new(configuration, identity, archive_store: archive_store).call
      end
      @mutex = Mutex.new
      @booted = false
      @shutdown = false
      @enabled = false
      @degraded = false
    end

    def boot(worker_index: nil)
      @mutex.synchronize do
        return self if @booted

        @booted = true
        start(worker_index)
      rescue StandardError => error
        degrade(error)
        cleanup
      end
      self
    end

    def shutdown
      @mutex.synchronize do
        return self if @shutdown

        @shutdown = true
        cleanup
      end
      self
    end

    def status
      writer_status = @writer&.status || {}
      archive_status = @archive_store&.status || {}
      emitter_status = @emitter&.status
      {
        enabled: @enabled,
        booted: @booted,
        shutdown: @shutdown,
        current_path: writer_status[:current_path],
        degraded: @degraded || writer_status[:degraded] || archive_status[:degraded],
        failure_count: writer_status.fetch(:failure_count) { archive_status.fetch(:failure_count, 0) },
        last_error: @last_error || emitter_status&.fetch(:last_write_error, nil) || writer_status[:last_error] || archive_status[:last_error],
        emitter: emitter_status
      }
    end

    private

    def start(worker_index)
      @configuration = ConfigurationFactory.from_env(@env)
      @enabled = @configuration.enabled
      return unless @enabled

      worker_id = worker_index.nil? ? nil : "worker-#{worker_index}"
      @identity = ProcessIdentity.resolve(service_instance_id: @configuration.service_instance_id, worker_id: worker_id, env: @env)
      @archive_store = @archive_builder.call(@configuration, @identity)
      @archive_store.recover!
      @writer = @writer_builder.call(@configuration, @identity, @archive_store)
      @emitter = Emitter.new(
        identity: @identity,
        writer: @writer,
        build_version: @configuration.build_version,
        environment: @configuration.environment
      )
      @file_logger = CanonicalLogger.new(StructuredFormatter.new(@emitter))
      @logger = @logger_provider.call
      raise ArgumentError, "Rails logger does not support broadcast sinks" unless @logger.respond_to?(:broadcast_to)

      @file_logger.level = @logger.level if @logger.respond_to?(:level) && @logger.level
      @logger.broadcast_to(@file_logger)
    end

    def cleanup
      @logger&.stop_broadcasting_to(@file_logger) if @file_logger && @logger&.respond_to?(:stop_broadcasting_to)
      if @file_logger
        @file_logger.close
      end
      @writer&.close
    rescue StandardError => error
      degrade(error)
      @writer&.close
    ensure
      @file_logger = nil
      @writer = nil if @shutdown
    end

    def degrade(error)
      @degraded = true
      @last_error = "#{error.class}: #{error.message}"
      report_to_console("API observability file output degraded: #{@last_error}")
    end

    def report_to_console(message)
      logger = @logger || @logger_provider.call
      logger.error(message)
    rescue StandardError
      warn(message)
    end
  end
end