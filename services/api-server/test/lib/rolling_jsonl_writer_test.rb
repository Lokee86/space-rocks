require "test_helper"
require "tmpdir"
require "zlib"

class RollingJsonlWriterTest < ActiveSupport::TestCase
  Identity = Observability::ProcessIdentity

  def configuration(root, **overrides)
    Observability::Configuration.new(**{
      enabled: true,
      log_root: root,
      service_instance_id: "550e8400-e29b-41d4-a716-446655440010",
      build_version: "test-build",
      environment: "test",
      segment_bytes: 12,
      segment_age: 60,
      retention_age: 3600,
      retention_bytes: 1024,
      compression: true
    }.merge(overrides))
  end

  def writer(root, configuration:, clock: -> { Time.now }, opener: nil)
    identity = Identity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 1)
    archive = Observability::ArchiveStore.new(configuration, identity, clock: clock)
    path = File.join(root, "active", "api-server-api-a-worker-1-pid-1.jsonl.open")
    FileUtils.mkdir_p(File.dirname(path))
    Observability::RollingJsonlWriter.new(path, configuration, archive, clock: clock, opener: opener)
  end

  test "size rotation never splits records and compresses the completed segment" do
    Dir.mktmpdir do |root|
      output = writer(root, configuration: configuration(root))
      output.write("first-line\n")
      output.write("second-line\n")
      archive = Dir.glob(File.join(root, "archive", "*.jsonl.gz")).sole
      assert_equal "first-line\n", Zlib::GzipReader.open(archive, &:read)
      assert_equal "second-line\n", File.read(output.path)
      output.close
    end
  end

  test "age rotation occurs before the next record" do
    Dir.mktmpdir do |root|
      now = Time.utc(2026, 7, 14, 12)
      clock = -> { now }
      config = configuration(root, segment_bytes: 1024, segment_age: 10, compression: false)
      output = writer(root, configuration: config, clock: clock)
      output.write("old\n")
      now += 10
      output.write("new\n")
      archive = Dir.glob(File.join(root, "archive", "*.jsonl")).sole
      assert_equal "old\n", File.read(archive)
      assert_equal "new\n", File.read(output.path)
      output.close
    end
  end

  test "close finalizes once and is idempotent" do
    Dir.mktmpdir do |root|
      output = writer(root, configuration: configuration(root, compression: false))
      output.write("record\n")
      2.times { output.close }
      assert output.closed?
      assert_equal 1, Dir.glob(File.join(root, "archive", "*.jsonl")).length
      assert_not File.exist?(output.path)
    end
  end

  test "write failure degrades without raising" do
    failing_io = Object.new
    failing_io.define_singleton_method(:binmode) {}
    failing_io.define_singleton_method(:sync=) { |_value| }
    failing_io.define_singleton_method(:size) { 0 }
    failing_io.define_singleton_method(:write) { |_value| raise IOError, "disk unavailable" }
    failing_io.define_singleton_method(:close) {}
    Dir.mktmpdir do |root|
      output = writer(root, configuration: configuration(root), opener: ->(_path) { failing_io })
      assert_nothing_raised { output.write("record\n") }
      assert output.status[:degraded]
      assert_match(/disk unavailable/, output.status[:last_error])
      output.close
    end
  end

  test "worker paths and active handles remain isolated" do
    Dir.mktmpdir do |root|
      config = configuration(root, segment_bytes: 1024, compression: false)
      writers = [1, 2].map do |index|
        identity = Identity.new(service_instance_id: "api-a", worker_id: "worker-#{index}", pid: 20 + index)
        archive = Observability::ArchiveStore.new(config, identity)
        Observability::WriterFactory.new(config, identity, archive_store: archive).call
      end
      writers.each_with_index { |output, index| output.write("worker-#{index + 1}\n") }
      assert_equal 2, writers.map(&:path).uniq.length
      assert_equal ["worker-1\n", "worker-2\n"], writers.map { |output| File.read(output.path) }
      writers.each(&:close)
    end
  end
end