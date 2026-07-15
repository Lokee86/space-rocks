require "test_helper"
require "tmpdir"

class WriterFactoryTest < ActiveSupport::TestCase
  test "derives unique active paths for workers and pids" do
    Dir.mktmpdir do |root|
      configuration = Observability::ConfigurationFactory.from_env("API_OBSERVABILITY_LOG_ROOT" => root)
      paths = [
        ["worker-1", 10],
        ["worker-2", 10],
        ["worker-1", 11]
      ].map do |worker_id, pid|
        identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: worker_id, pid: pid)
        Observability::WriterFactory.new(configuration, identity, writer: ->(path) { path }).call
      end

      assert_equal 3, paths.uniq.length
      assert paths.all? { |path| path.end_with?(".jsonl.open") }
    end
  end

  test "delegates writer construction with the derived path" do
    Dir.mktmpdir do |root|
      configuration = Observability::ConfigurationFactory.from_env("API_OBSERVABILITY_LOG_ROOT" => root)
      identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: "worker/1", pid: 10)
      received = nil
      result = Observability::WriterFactory.new(configuration, identity, writer: ->(path) { received = path; :writer }).call
      assert_equal :writer, result
      assert_equal File.join(root, "active", "api-server-api-a-worker_1-pid-10.jsonl.open"), received
    end
  end

  test "builds the concrete rolling writer" do
    Dir.mktmpdir do |root|
      configuration = Observability::ConfigurationFactory.from_env(
        "API_OBSERVABILITY_LOG_ROOT" => root,
        "API_OBSERVABILITY_SEGMENT_BYTES" => "1024"
      )
      identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 10)
      archive = Observability::ArchiveStore.new(configuration, identity)
      writer = Observability::WriterFactory.new(configuration, identity, archive_store: archive).call
      assert_instance_of Observability::RollingJsonlWriter, writer
      assert File.exist?(writer.path)
      writer.close
    end
  end

  test "refuses to overwrite a surviving active segment" do
    Dir.mktmpdir do |root|
      configuration = Observability::ConfigurationFactory.from_env("API_OBSERVABILITY_LOG_ROOT" => root)
      identity = Observability::ProcessIdentity.new(service_instance_id: "api-a", worker_id: "worker-1", pid: 10)
      path = File.join(root, "active", "api-server-api-a-worker-1-pid-10.jsonl.open")
      FileUtils.mkdir_p(File.dirname(path))
      File.write(path, "recover me")
      assert_raises(IOError) do
        Observability::WriterFactory.new(configuration, identity, writer: ->(_path) { flunk }).call
      end
      assert_equal "recover me", File.read(path)
    end
  end
end