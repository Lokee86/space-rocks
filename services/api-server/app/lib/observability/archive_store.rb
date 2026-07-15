require "fileutils"
require "zlib"
require_relative "retention_lock"

module Observability
  class ArchiveStore
    attr_reader :last_error, :failure_count

    def initialize(configuration, identity, clock: -> { Time.now }, lock: nil)
      @configuration = configuration
      @identity = identity
      @clock = clock
      @lock = lock || RetentionLock.new(configuration.log_root)
      @degraded = false
      @failure_count = 0
    end

    def recover!
      stale_active_paths.each { |path| finalize(path, retain: false) }
      retain!
    rescue StandardError => error
      degrade(error)
      false
    end

    def finalize(active_path, retain: true)
      return unless File.exist?(active_path)

      FileUtils.mkdir_p(archive_directory)
      archive_path = unique_archive_path(File.basename(active_path).delete_suffix(".open"))
      File.rename(active_path, archive_path)
      completed_path = @configuration.compression ? compress(archive_path) : archive_path
      retain! if retain
      completed_path
    rescue StandardError => error
      degrade(error)
      nil
    end

    def retain!
      @lock.call do
        expired_before = @clock.call - @configuration.retention_age
        archive_paths.select { |path| File.mtime(path) < expired_before }.each { |path| remove(path) }
        remaining = archive_paths.sort_by { |path| [File.mtime(path), path] }
        total_bytes = remaining.sum { |path| File.size(path) }
        remaining.each do |path|
          break if total_bytes <= @configuration.retention_bytes

          bytes = File.size(path)
          total_bytes -= bytes if remove(path)
        end
      end
      !@degraded
    rescue StandardError => error
      degrade(error)
      false
    end

    def status
      { degraded: @degraded, failure_count: @failure_count, last_error: @last_error }
    end

    private

    def stale_active_paths
      pattern = "api-server-#{component(@identity.service_instance_id)}-#{component(@identity.worker_id)}-pid-*.jsonl.open"
      Dir.glob(File.join(@configuration.log_root.to_s, "active", pattern)).sort
    end

    def archive_paths
      Dir.glob([File.join(archive_directory, "*.jsonl"), File.join(archive_directory, "*.jsonl.gz")]).uniq
    end

    def archive_directory
      File.join(@configuration.log_root.to_s, "archive")
    end

    def unique_archive_path(filename)
      stem = filename.delete_suffix(".jsonl")
      index = 0
      loop do
        suffix = index.zero? ? "" : "-#{index}"
        candidate = File.join(archive_directory, "#{stem}#{suffix}.jsonl")
        return candidate unless File.exist?(candidate) || File.exist?("#{candidate}.gz")

        index += 1
      end
    end

    def compress(source_path)
      compressed_path = "#{source_path}.gz"
      temporary_path = "#{compressed_path}.tmp-#{Process.pid}"
      Zlib::GzipWriter.open(temporary_path) do |gzip|
        File.open(source_path, "rb") { |source| IO.copy_stream(source, gzip) }
      end
      File.rename(temporary_path, compressed_path)
      File.delete(source_path)
      compressed_path
    rescue StandardError => error
      FileUtils.rm_f(temporary_path) if temporary_path
      degrade(error)
      source_path
    end

    def remove(path)
      File.delete(path)
      true
    rescue StandardError => error
      degrade(error)
      false
    end

    def component(value)
      value.to_s.gsub(/[^0-9A-Za-z_.-]/, "_")
    end

    def degrade(error)
      @degraded = true
      @failure_count += 1
      @last_error = "#{error.class}: #{error.message}"
    end
  end
end