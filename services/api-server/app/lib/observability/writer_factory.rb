require "fileutils"

module Observability
  class WriterFactory
    def initialize(configuration, identity, writer:)
      @configuration = configuration
      @identity = identity
      @writer = writer
    end

    def call
      path = active_path
      FileUtils.mkdir_p(File.dirname(path))
      @writer.call(path)
    end

    private

    def active_path
      root = @configuration.log_root.to_s
      filename = "api-server-#{component(@identity.service_instance_id)}-#{component(@identity.worker_id)}-pid-#{@identity.pid}.jsonl.open"
      File.join(root, "active", filename)
    end

    def component(value)
      value.to_s.gsub(/[^0-9A-Za-z_.-]/, "_")
    end
  end
end