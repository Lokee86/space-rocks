require "thread"

module Observability
  class RollingJsonlWriter
    attr_reader :path, :last_error, :failure_count

    def initialize(path, configuration, archive_store, clock: -> { Time.now }, opener: nil)
      @path = path.to_s
      @configuration = configuration
      @archive_store = archive_store
      @clock = clock
      @opener = opener || ->(active_path) { File.open(active_path, "ab") }
      @mutex = Mutex.new
      @closed = false
      @degraded = false
      @failure_count = 0
      open_active
    end

    def write(record)
      payload = record.to_s
      @mutex.synchronize do
        return payload.bytesize if @closed || @degraded

        rotate if rotation_due?(payload.bytesize)
        return payload.bytesize if @degraded

        @io.write(payload)
        @bytes += payload.bytesize
      rescue StandardError => error
        degrade(error)
      end
      payload.bytesize
    end

    def flush
      @mutex.synchronize do
        @io&.flush unless @closed || @degraded
      rescue StandardError => error
        degrade(error)
      end
      self
    end

    def close
      @mutex.synchronize do
        return self if @closed

        finish_active
        @closed = true
      end
      self
    end

    def closed?
      @closed
    end

    def status
      archive_status = @archive_store.status
      {
        enabled: !@closed && !@degraded,
        current_path: @io ? @path : nil,
        degraded: @degraded || archive_status[:degraded],
        failure_count: @failure_count + archive_status.fetch(:failure_count, 0),
        last_error: @last_error || archive_status[:last_error]
      }
    end

    private

    def open_active
      @io = @opener.call(@path)
      @io.binmode if @io.respond_to?(:binmode)
      @io.sync = true if @io.respond_to?(:sync=)
      @bytes = @io.respond_to?(:size) ? @io.size : File.size(@path)
      @segment_started_at = @clock.call
    end

    def rotation_due?(incoming_bytes)
      return false if @bytes.zero?

      @bytes + incoming_bytes > @configuration.segment_bytes || @clock.call - @segment_started_at >= @configuration.segment_age
    end

    def rotate
      finish_active
      return degrade(RuntimeError.new("failed to finalize active observability segment")) if File.exist?(@path)

      open_active
    end

    def finish_active
      return unless @io

      @io.flush
      @io.close
      @io = nil
      @archive_store.finalize(@path) if File.exist?(@path)
    rescue StandardError => error
      degrade(error)
    ensure
      @io = nil if @io&.closed?
    end

    def degrade(error)
      @degraded = true
      @failure_count += 1
      @last_error = "#{error.class}: #{error.message}"
      @io&.close
      @io = nil
      nil
    rescue StandardError
      @io = nil
      nil
    end
  end
end