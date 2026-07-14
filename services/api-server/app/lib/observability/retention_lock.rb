require "fileutils"

module Observability
  class RetentionLock
    def initialize(log_root, opener: ->(path) { File.open(path, "a+") })
      @log_root = log_root.to_s
      @opener = opener
    end

    def call
      FileUtils.mkdir_p(state_directory)
      lock_file = @opener.call(lock_path)
      lock_file.flock(File::LOCK_EX)
      yield
    ensure
      lock_file&.flock(File::LOCK_UN)
      lock_file&.close
    end

    def lock_path
      File.join(state_directory, "retention.lock")
    end

    private

    def state_directory
      File.join(@log_root, "state")
    end
  end
end