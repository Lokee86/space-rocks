require "test_helper"
require "json"
require "stringio"
require "tmpdir"
require "zlib"

class WorkerRuntimeTest < ActiveSupport::TestCase
  UUID = "550e8400-e29b-41d4-a716-446655440010"

  def broadcast_logger(console_output)
    ActiveSupport::BroadcastLogger.new(Logger.new(console_output))
  end

  def enabled_env(root)
    {
      "API_OBSERVABILITY_ENABLED" => "true",
      "API_OBSERVABILITY_LOG_ROOT" => root,
      "API_SERVICE_INSTANCE_ID" => UUID,
      "BUILD_VERSION" => "test-build",
      "RAILS_ENV" => "test",
      "API_OBSERVABILITY_SEGMENT_BYTES" => "1024"
    }
  end

  test "boot opens a worker-owned emitter and shutdown archives emitted records" do
    Dir.mktmpdir do |root|
      console = StringIO.new
      logger = broadcast_logger(console)
      runtime = Observability::WorkerRuntime.new(env: enabled_env(root), logger_provider: -> { logger })
      runtime.boot(worker_index: 3)
      active_path = runtime.status[:current_path]
      result = runtime.emit(event: "build_version_loaded", message: "hello")

      assert result[:accepted]
      assert File.exist?(active_path)
      payload = JSON.parse(File.read(active_path))
      assert_equal "api-server", payload["service"]
      assert_equal UUID, payload["service_instance_id"]
      assert_equal "worker-3", payload["worker_id"]
      assert_equal "test-build", payload["build_version"]
      assert_equal "test", payload["environment"]
      assert_empty console.string

      2.times { runtime.shutdown }
      archived = Dir.glob(File.join(root, "archive", "*.jsonl.gz")).sole
      assert_includes Zlib::GzipReader.open(archived, &:read), "hello"
      logger.info("console-only")
      assert_includes console.string, "console-only"
      assert runtime.status[:shutdown]
    end
  end

  test "repeated boot constructs only one worker writer" do
    Dir.mktmpdir do |root|
      logger = broadcast_logger(StringIO.new)
      builds = 0
      builder = lambda do |configuration, identity, archive|
        builds += 1
        Observability::WriterFactory.new(configuration, identity, archive_store: archive).call
      end
      runtime = Observability::WorkerRuntime.new(
        env: enabled_env(root), logger_provider: -> { logger }, writer_builder: builder
      )
      2.times { runtime.boot(worker_index: 1) }
      assert_equal 1, builds
      runtime.shutdown
    end
  end

  test "partial startup failure degrades to bounded stderr without using the Rails logger" do
    Dir.mktmpdir do |root|
      console = StringIO.new
      logger = broadcast_logger(console)
      warning = StringIO.new
      runtime = Observability::WorkerRuntime.new(
        env: enabled_env(root),
        logger_provider: -> { logger },
        writer_builder: ->(*) { raise IOError, "read-only volume" },
        warning_io: warning
      )
      assert_nothing_raised { runtime.boot(worker_index: 2) }
      assert runtime.status[:degraded]
      assert_match(/read-only volume/, runtime.status[:last_error])
      assert_match(/read-only volume/, warning.string)
      assert_empty console.string
      assert_nothing_raised { logger.info("request still served") }
      assert_includes console.string, "request still served"
      assert_nothing_raised { runtime.shutdown }
    end
  end

  test "invalid file identity degrades without disabling the Rails logger" do
    console = StringIO.new
    logger = broadcast_logger(console)
    warning = StringIO.new
    runtime = Observability::WorkerRuntime.new(
      env: { "API_OBSERVABILITY_ENABLED" => "true", "API_SERVICE_INSTANCE_ID" => "invalid" },
      logger_provider: -> { logger },
      warning_io: warning
    )
    assert_nothing_raised { runtime.boot }
    assert runtime.status[:degraded]
    assert_match(/must be a UUID/, warning.string)
    assert_empty console.string
    logger.info("request still served")
    assert_includes console.string, "request still served"
    runtime.shutdown
  end

  test "disabled configuration opens no file sink and emission is a no-op" do
    console = StringIO.new
    logger = broadcast_logger(console)
    runtime = Observability::WorkerRuntime.new(env: {}, logger_provider: -> { logger })
    runtime.boot
    assert_not runtime.status[:enabled]
    assert_not runtime.status[:degraded]
    assert_nil runtime.status[:current_path]
    assert runtime.emit(event: "build_version_loaded")[:disabled]
    runtime.shutdown
  end
end
