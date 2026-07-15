require "logger"
require "thread"
require_relative "archive_store"
require_relative "configuration_factory"
require_relative "emitter"
require_relative "process_identity"
require_relative "writer_factory"

module Observability
  class WorkerRuntime
    def initialize(env: ENV, logger_provider: nil, archive_builder: nil, writer_builder: nil, warning_io: $stderr)
      @env = env
      @warning_io = warning_io
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
        environment: @configuration.environment,
        warning_io: @warning_io
      )
    end

    def emit(event:, message: nil, context: {}, fields: {})
      return disabled_emission unless @enabled && @emitter

      @emitter.emit(event: event, message: message, context: context, fields: fields)
    rescue StandardError
      disabled_emission
    end

    def emitter_status
      @emitter&.status || { enabled: @enabled, disabled: !@enabled }
    end

    public :emit, :emitter_status

    private

    def cleanup
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
      @warning_io.puts(message.to_s.byteslice(0, 512))
    rescue StandardError
      nil
    end

    def disabled_emission
      { accepted: false, disabled: true, redacted: false, rejection_code: nil, rejected_key: nil, write_failed: false }
    end
  end
end