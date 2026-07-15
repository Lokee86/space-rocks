require "test_helper"
require "tmpdir"

class RetentionLockTest < ActiveSupport::TestCase
  test "creates and releases the retention lock around the block" do
    Dir.mktmpdir do |root|
      lock = Observability::RetentionLock.new(root)
      inside = false

      result = lock.call do
        inside = File.exist?(lock.lock_path)
        :completed
      end

      assert_equal :completed, result
      assert inside
      assert File.exist?(lock.lock_path)
      File.open(lock.lock_path, "a+") do |file|
        assert file.flock(File::LOCK_EX | File::LOCK_NB)
      end
    end
  end

  test "uses the injected file-opening boundary" do
    opened_path = nil
    opener = ->(path) { opened_path = path; File.open(path, "a+") }

    Dir.mktmpdir do |root|
      lock = Observability::RetentionLock.new(root, opener: opener)
      lock.call { :done }

      assert_equal lock.lock_path, opened_path
    end
  end
end