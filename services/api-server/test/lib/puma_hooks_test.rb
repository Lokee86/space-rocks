require "test_helper"

class PumaHooksTest < ActiveSupport::TestCase
  class FakePumaConfig
    attr_reader :boot_callback, :shutdown_callback

    def before_worker_boot(&block)
      @boot_callback = block
    end

    def before_worker_shutdown(&block)
      @shutdown_callback = block
    end
  end

  test "registers boot and shutdown callbacks" do
    config = FakePumaConfig.new

    Observability::PumaHooks.new.register(config)

    assert config.boot_callback
    assert config.shutdown_callback
  end

  test "forwards the worker index after worker boot" do
    config = FakePumaConfig.new
    boot_indexes = []

    Observability::PumaHooks.new(worker_boot: ->(index) { boot_indexes << index }).register(config)

    config.boot_callback.call(4)

    assert_equal [4], boot_indexes
  end

  test "forwards the worker index before worker shutdown" do
    config = FakePumaConfig.new
    shutdown_indexes = []

    Observability::PumaHooks.new(worker_shutdown: ->(index) { shutdown_indexes << index }).register(config)

    config.shutdown_callback.call(5)

    assert_equal [5], shutdown_indexes
  end

  test "defaults are safe no-ops" do
    config = FakePumaConfig.new

    assert_nothing_raised do
      Observability::PumaHooks.new.register(config)
      config.boot_callback.call(0)
      config.shutdown_callback.call(0)
    end
  end
end