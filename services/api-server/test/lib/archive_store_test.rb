require "test_helper"
require "tmpdir"
require "zlib"

class ArchiveStoreTest < ActiveSupport::TestCase
  Identity = Observability::ProcessIdentity

  def configuration(root, **overrides)
    Observability::Configuration.new(**{
      enabled: true,
      log_root: root,
      service_instance_id: "550e8400-e29b-41d4-a716-446655440010",
      build_version: "test-build",
      environment: "test",
      segment_bytes: 100,
      segment_age: 60,
      retention_age: 100,
      retention_bytes: 100,
      compression: true
    }.merge(overrides))
  end

  test "recovers only stale active segments owned by the worker and compresses them" do
    Dir.mktmpdir do |root|
      active = File.join(root, "active")
      FileUtils.mkdir_p(active)
      instance = "550e8400-e29b-41d4-a716-446655440010"
      owned = File.join(active, "api-server-#{instance}-worker-1-pid-10.jsonl.open")
      other = File.join(active, "api-server-#{instance}-worker-2-pid-11.jsonl.open")
      File.binwrite(owned, "{\"owned\":true}\n")
      File.binwrite(other, "{\"other\":true}\n")
      store = Observability::ArchiveStore.new(configuration(root), Identity.new(service_instance_id: instance, worker_id: "worker-1", pid: 12))
      assert store.recover!
      archive = Dir.glob(File.join(root, "archive", "*.jsonl.gz")).sole
      assert_equal "{\"owned\":true}\n", Zlib::GzipReader.open(archive, &:read)
      assert_not File.exist?(owned)
      assert File.exist?(other)
    end
  end

  test "retention deletes expired then oldest archives under the lock" do
    Dir.mktmpdir do |root|
      archive = File.join(root, "archive")
      FileUtils.mkdir_p(archive)
      old = File.join(archive, "old.jsonl.gz")
      first = File.join(archive, "first.jsonl.gz")
      second = File.join(archive, "second.jsonl.gz")
      [old, first, second].each { |path| File.binwrite(path, "x" * 6) }
      now = Time.utc(2026, 7, 14, 12)
      File.utime(now - 200, now - 200, old)
      File.utime(now - 20, now - 20, first)
      File.utime(now - 10, now - 10, second)
      calls = 0
      lock = Object.new
      lock.define_singleton_method(:call) { |&block| calls += 1; block.call }
      store = Observability::ArchiveStore.new(
        configuration(root, retention_bytes: 6),
        Identity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 1),
        clock: -> { now }, lock: lock
      )
      assert store.retain!
      assert_equal 1, calls
      assert_not File.exist?(old)
      assert_not File.exist?(first)
      assert File.exist?(second)
    end
  end

  test "does not overwrite an existing archive" do
    Dir.mktmpdir do |root|
      active = File.join(root, "active", "api-server-api-a-worker-1-pid-1.jsonl.open")
      archive = File.join(root, "archive", "api-server-api-a-worker-1-pid-1.jsonl")
      FileUtils.mkdir_p(File.dirname(active))
      FileUtils.mkdir_p(File.dirname(archive))
      File.write(active, "new")
      File.write(archive, "old")
      store = Observability::ArchiveStore.new(configuration(root, compression: false), Identity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 1))
      completed = store.finalize(active)
      assert_equal "old", File.read(archive)
      assert_equal "new", File.read(completed)
      assert_not_equal archive, completed
    end
  end

  test "compression failure preserves the completed plain archive" do
    Dir.mktmpdir do |root|
      active = File.join(root, "active", "api-server-api-a-worker-1-pid-1.jsonl.open")
      FileUtils.mkdir_p(File.dirname(active))
      File.write(active, "recoverable")
      store = Observability::ArchiveStore.new(configuration(root), Identity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 1))
      completed = with_singleton_method_stub(Zlib::GzipWriter, :open, ->(*) { raise IOError, "compression failed" }) do
        store.finalize(active)
      end
      assert completed.end_with?(".jsonl")
      assert_equal "recoverable", File.read(completed)
      assert store.status[:degraded]
      assert_match(/compression failed/, store.status[:last_error])
    end
  end
end